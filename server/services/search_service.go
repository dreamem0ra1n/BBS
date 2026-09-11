package services

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"

	"bbs-go/cache"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/config"
	pkghtml "bbs-go/pkg/html"
	pkgmarkdown "bbs-go/pkg/markdown"
	"bbs-go/pkg/meili"
	"bbs-go/repositories"
)

var SearchService = newSearchService()

const (
	// 文档 ID 只能用字母数字、- 和 _，不能用冒号等字符，
	// 否则 Meilisearch 会判定 invalid_document_id 并整批拒绝写入。
	searchDocIdNewPrefix = "new-"
	searchDocIdOldPrefix = "old-"

	// 索引进正文的最大字符数，避免旧站超长正文把索引撑大
	searchContentMaxRunes = 2000

	searchPermissionModeFilter = "filter"
	searchPermissionModeMark   = "mark"
)

// SearchParams 搜索参数。
// User 用于权限判断，匿名时为 nil。
type SearchParams struct {
	Keyword  string
	NodeId   int64
	IsOldBBS bool
	Page     int
	Limit    int
	User     *model.User
}

type searchService struct {
	client         *meili.Client
	index          string
	enabled        bool
	permissionMode string
	maxLimit       int
	maxPage        int
	maxScan        int
	syncOnBoot     bool
	syncOldBBS     bool

	nodeCacheMu sync.Mutex
	nodeCache   []model.TopicNode
	nodeCacheAt time.Time
}

func newSearchService() *searchService {
	// 注意：这里不能读取 config.Instance，services 包的变量初始化早于 main 的 init()。
	// 配置在 Init() 里统一读取。
	return &searchService{}
}

// Init 初始化搜索引擎：建索引、写索引配置，并按需触发全量重建。
// 任何一步失败都只记录日志并且不影响主流程，搜索会自动回退到数据库。
func (s *searchService) Init() {
	if config.Instance == nil {
		return
	}
	cfg := config.Instance.Search
	s.index = cfg.Index
	s.permissionMode = cfg.PermissionMode
	s.maxLimit = cfg.MaxLimit
	s.maxPage = cfg.MaxPage
	s.maxScan = cfg.MaxScan
	s.syncOnBoot = cfg.SyncOnBoot
	s.syncOldBBS = cfg.SyncOldBBS
	s.enabled = cfg.Enabled

	if !s.enabled {
		logrus.Info("search: meilisearch disabled, use database like as fallback")
		return
	}

	s.client = meili.NewClient(cfg.Url, cfg.ApiKey, time.Duration(cfg.TimeoutMs)*time.Millisecond)
	if err := s.client.Health(); err != nil {
		// 不关闭搜索：运行期每次查询失败都会回退到数据库，等服务恢复后自动可用。
		logrus.Warn("search: meilisearch health check failed, fallback to database like until it recovers: ", err)
		return
	}
	if err := s.client.EnsureIndex(s.index, "id"); err != nil {
		logrus.Error("search: ensure index failed: ", err)
		return
	}
	if err := s.client.UpdateSettings(s.index, meili.Settings{
		// 越靠前权重越高：标题 > 标签 > 正文 > 作者/节点
		SearchableAttributes: []string{"title", "tags", "content", "authorName", "nodeName"},
		FilterableAttributes: []string{"status", "isOldBBS", "nodeId", "recommend", "accessLv", "userId", "createTime"},
		SortableAttributes:   []string{"createTime", "lastCommentTime", "viewCount", "commentCount", "likeCount"},
		RankingRules:         []string{"words", "typo", "proximity", "attribute", "sort", "exactness"},
		Pagination:           &meili.Pagination{MaxTotalHits: 2000},
	}); err != nil {
		logrus.Error("search: update index settings failed: ", err)
		return
	}
	logrus.Info("search: meilisearch initialized, index=", s.index)

	if s.syncOnBoot {
		go s.ReindexAll()
		if s.syncOldBBS {
			go s.ReindexOldBBS()
		}
	}
}

func (s *searchService) meiliEnabled() bool {
	return s.enabled && s.client != nil
}

func (s *searchService) limitCap() int {
	if s.maxLimit <= 0 {
		return 50
	}
	return s.maxLimit
}

func (s *searchService) pageCap() int {
	if s.maxPage <= 0 {
		return 100
	}
	return s.maxPage
}

func (s *searchService) scanCap() int {
	if s.maxScan <= 0 {
		return 1000
	}
	return s.maxScan
}

// clampPaging 限制 page/limit 的取值，避免 limit 被放大后打满数据库或搜索引擎。
func (s *searchService) clampPaging(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > s.limitCap() {
		limit = s.limitCap()
	}
	return page, limit
}

// SearchTopics 搜索主题。
// 优先走 Meilisearch，失败或未启用时回退到数据库 LIKE。
func (s *searchService) SearchTopics(p SearchParams) ([]model.Topic, *sqls.Paging) {
	p.Page, p.Limit = s.clampPaging(p.Page, p.Limit)
	p.Keyword = strings.TrimSpace(p.Keyword)

	// 未登录访客不可使用搜索
	if p.User == nil {
		return nil, &sqls.Paging{Page: p.Page, Limit: p.Limit, Total: 0}
	}
	// 页码越界：直接返回空，避免深分页全索引扫描
	if p.Page > s.pageCap() {
		return nil, &sqls.Paging{Page: p.Page, Limit: p.Limit, Total: 0}
	}
	// 空关键词会导致全表/全索引匹配，直接短路
	if p.Keyword == "" {
		return nil, &sqls.Paging{Page: p.Page, Limit: p.Limit, Total: 0}
	}

	if s.meiliEnabled() {
		topics, paging, err := s.searchViaMeili(p)
		if err == nil {
			return topics, paging
		}
		logrus.Warn("search: meilisearch query failed, fallback to database like: ", err)
	}
	return s.searchViaSql(p)
}

// buildFilter 构造 Meilisearch 过滤表达式。
// 这里始终带 status = 0，保证已删除主题不会出现在搜索结果里。
func (s *searchService) buildFilter(p SearchParams) string {
	parts := []string{
		"status = " + strconv.Itoa(constants.StatusOk),
		"isOldBBS = " + strconv.FormatBool(p.IsOldBBS),
	}
	if !p.IsOldBBS {
		if p.NodeId == -1 {
			parts = append(parts, "recommend = true")
		} else if p.NodeId > 0 {
			parts = append(parts, "nodeId = "+strconv.FormatInt(p.NodeId, 10))
		}
	}
	// 权限条件下推到引擎，保证 estimatedTotalHits 就是“权限过滤后的总数”，
	// 同时让分页 offset 直接可用；Go 侧仍会再校验一次作为兜底。
	if permissionFilter := s.buildPermissionFilter(p); permissionFilter != "" {
		parts = append(parts, permissionFilter)
	}
	return strings.Join(parts, " AND ")
}

// buildPermissionFilter 把 model.UserCanAccessTopic 的判定翻译成 Meilisearch 过滤表达式。
// 说明：
//   - 新站权限按“节点 + 用户在该节点的部门角色 read_lv”决定，所以这里按节点逐个展开；
//   - 不在节点表里的 nodeId 使用默认角色（ReadLv=1），与 GetRoleByArg 找不到匹配时的行为一致；
//   - 作者本人始终能看到自己的主题（与帖子详情页一致）。
func (s *searchService) buildPermissionFilter(p SearchParams) string {
	if s.permissionMode == searchPermissionModeMark || p.User == nil {
		return ""
	}
	return buildTopicPermissionFilter(p.User, s.nodes(), p.IsOldBBS)
}

// buildTopicPermissionFilter 是 buildPermissionFilter 的纯函数部分，便于单测。
func buildTopicPermissionFilter(user *model.User, nodes []model.TopicNode, isOldBBS bool) string {
	if user == nil {
		return ""
	}
	// 站长/高管看全部
	if user.IsMasterUser() {
		return ""
	}

	if isOldBBS {
		// 旧站：readperm 为 0 的帖子所有人可见，否则只有 oldbbs_readall 角色可见
		if user.HasAnyRole(model.OLDBBSUser_NAME) {
			return ""
		}
		return "accessLv = 0"
	}

	// 按 read_lv 分组节点：
	//   read_lv == 0  -> 该节点下所有帖子都可读
	//   read_lv >= 1  -> 要求 access_lv != 0 且 access_lv <= read_lv
	// 分组后过滤表达式长度只与角色种类数有关，不随节点数量线性增长。
	unrestricted := make([]string, 0)
	restricted := make(map[int][]string)
	nodeIds := make([]string, 0)
	for _, node := range nodes {
		nodeIdStr := strconv.FormatInt(node.Id, 10)
		nodeIds = append(nodeIds, nodeIdStr)
		role, err := user.GetRoleByArg(node.Id)
		if err != nil {
			// 与 UserCanAccessTopic 一致：角色异常视为不可访问
			continue
		}
		authUnit, err := model.GetAuthUnit(role, int(node.Id))
		if err != nil {
			continue
		}
		if authUnit.ReadLv == 0 {
			unrestricted = append(unrestricted, nodeIdStr)
		} else if authUnit.ReadLv > 0 {
			restricted[authUnit.ReadLv] = append(restricted[authUnit.ReadLv], nodeIdStr)
		}
	}

	parts := []string{"userId = " + strconv.FormatInt(user.Id, 10)}
	if len(unrestricted) > 0 {
		parts = append(parts, "nodeId IN ["+strings.Join(unrestricted, ", ")+"]")
	}
	readLvs := make([]int, 0, len(restricted))
	for readLv := range restricted {
		readLvs = append(readLvs, readLv)
	}
	sort.Ints(readLvs)
	for _, readLv := range readLvs {
		parts = append(parts, "(nodeId IN ["+strings.Join(restricted[readLv], ", ")+
			"] AND accessLv != 0 AND accessLv <= "+strconv.Itoa(readLv)+")")
	}

	defaultReadLv := 1
	if authUnit, err := model.GetAuthUnit(model.DefaultUser_NAME, model.Default_SECTION); err == nil {
		defaultReadLv = authUnit.ReadLv
	}
	// 不在节点表里的 nodeId 使用默认角色（与 GetRoleByArg 找不到匹配时一致）
	defaultClause := "(accessLv != 0 AND accessLv <= " + strconv.Itoa(defaultReadLv) + ")"
	if len(nodeIds) > 0 {
		defaultClause = "(nodeId NOT IN [" + strings.Join(nodeIds, ", ") + "] AND accessLv != 0 AND accessLv <= " +
			strconv.Itoa(defaultReadLv) + ")"
	}
	parts = append(parts, defaultClause)

	return "(" + strings.Join(parts, " OR ") + ")"
}

// nodes 带 1 分钟缓存地返回节点列表，避免每次搜索都查一次节点表。
func (s *searchService) nodes() []model.TopicNode {
	s.nodeCacheMu.Lock()
	defer s.nodeCacheMu.Unlock()
	if s.nodeCache != nil && time.Since(s.nodeCacheAt) < time.Minute {
		return s.nodeCache
	}
	nodes := TopicNodeService.GetNodes()
	if nodes == nil {
		nodes = []model.TopicNode{}
	}
	s.nodeCache = nodes
	s.nodeCacheAt = time.Now()
	return nodes
}

func (s *searchService) searchViaMeili(p SearchParams) ([]model.Topic, *sqls.Paging, error) {
	skip := (p.Page - 1) * p.Limit
	filter := s.buildFilter(p)

	// 每次多取一些候选，因为权限过滤会消耗掉一部分
	batchSize := p.Limit * 2
	if batchSize < 50 {
		batchSize = 50
	}
	if batchSize > 200 {
		batchSize = 200
	}

	var (
		results        []model.Topic
		offset         int
		scanned        int
		visibleSkipped int
		total          int64
	)
	scanCap := s.scanCap()
	for len(results) < p.Limit && scanned < scanCap {
		resp, err := s.client.Search(s.index, meili.SearchRequest{
			Q:      p.Keyword,
			Filter: filter,
			Limit:  batchSize,
			Offset: offset,
			// 相关度优先，时间倒序作为并列时的次序
			Sort:                 []string{"createTime:desc"},
			AttributesToRetrieve: []string{"id"},
		})
		if err != nil {
			return nil, nil, err
		}
		total = resp.EstimatedTotalHits
		if len(resp.Hits) == 0 {
			break
		}

		ids := make([]int64, 0, len(resp.Hits))
		for _, hit := range resp.Hits {
			if _, id, ok := parseSearchDocId(hit.ID); ok {
				ids = append(ids, id)
			}
		}
		topics := s.hydrate(p.IsOldBBS, ids)
		for i := range topics {
			topic := &topics[i]
			if !s.canSee(p.User, topic) {
				continue
			}
			if visibleSkipped < skip {
				visibleSkipped++
				continue
			}
			results = append(results, *topic)
			if len(results) >= p.Limit {
				break
			}
		}

		offset += len(resp.Hits)
		scanned += len(resp.Hits)
		if len(resp.Hits) < batchSize {
			break
		}
		if total > 0 && int64(offset) >= total {
			break
		}
	}

	return results, &sqls.Paging{Page: p.Page, Limit: p.Limit, Total: total}, nil
}

// searchViaSql 数据库回退实现。
// 这里同样带 status = 0、LIKE 通配符转义和权限过滤，并且总数只统计“当前用户可见”的结果，
// 与搜索引擎路径保持一致。扫描条数受 MaxScan 限制，达到上限时总数为可见结果的下界。
func (s *searchService) searchViaSql(p SearchParams) ([]model.Topic, *sqls.Paging) {
	skip := (p.Page - 1) * p.Limit
	batchSize := 500
	scanCap := s.scanCap()

	var (
		results      = make([]model.Topic, 0, p.Limit)
		visibleTotal int64
		scanned      int
		offset       int
	)
	for scanned < scanCap {
		var batch []model.Topic
		if p.IsOldBBS {
			batch = OldBBSService.GetTopicsByKeywordOffset(p.Keyword, offset, batchSize)
		} else {
			batch = s.searchNewTopicsBySql(p.Keyword, p.NodeId, offset, batchSize)
		}
		if len(batch) == 0 {
			break
		}
		for i := range batch {
			topic := &batch[i]
			if !s.canSee(p.User, topic) {
				continue
			}
			visibleTotal++
			if int(visibleTotal) <= skip {
				continue
			}
			if len(results) < p.Limit {
				results = append(results, *topic)
			}
		}
		offset += len(batch)
		scanned += len(batch)
		if len(batch) < batchSize {
			break
		}
	}
	return results, &sqls.Paging{Page: p.Page, Limit: p.Limit, Total: visibleTotal}
}

func (s *searchService) searchNewTopicsBySql(keyword string, nodeId int64, offset, limit int) []model.Topic {
	pattern := "%" + escapeLike(keyword) + "%"
	// 显式声明转义符，避免关键词里的 % / _ 被当成通配符
	where := `status = 0 AND title LIKE ? ESCAPE '\\'`
	args := []interface{}{pattern}
	if nodeId == -1 {
		where += " AND recommend = 1"
	} else if nodeId > 0 {
		where += " AND node_id = ?"
		args = append(args, nodeId)
	}

	queryArgs := make([]interface{}, 0, len(args)+2)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, limit, offset)
	return repositories.TopicRepository.FindBySql(sqls.DB(),
		"SELECT * FROM t_topic WHERE "+where+" ORDER BY create_time DESC LIMIT ? OFFSET ?", queryArgs...)
}

// canSee 判断当前用户能否在搜索结果中看到该主题。
// 复用 model.UserCanAccessTopic，保证与帖子详情页的权限判断完全一致；
// 旧站 readperm、新站按部门角色的 read_lv 都由它处理。
func (s *searchService) canSee(user *model.User, topic *model.Topic) bool {
	if s.permissionMode == searchPermissionModeMark {
		return true
	}
	if topic.Status != constants.StatusOk {
		return false
	}
	if model.UserCanAccessTopic(user, topic) {
		return true
	}
	// 与详情页一致：作者本人始终可见自己的主题
	return user != nil && !topic.IsOldBBS && topic.UserId == user.Id
}

func (s *searchService) hydrate(isOldBBS bool, ids []int64) []model.Topic {
	if len(ids) == 0 {
		return nil
	}
	if isOldBBS {
		return orderTopics(ids, OldBBSService.GetTopicsByIds(ids))
	}
	return orderTopics(ids, TopicService.GetTopicInIds(ids))
}

// orderTopics 按搜索引擎返回的顺序重新排列回表结果。
func orderTopics(ids []int64, topicsMap map[int64]model.Topic) []model.Topic {
	if len(topicsMap) == 0 {
		return nil
	}
	topics := make([]model.Topic, 0, len(ids))
	for _, id := range ids {
		if topic, ok := topicsMap[id]; ok {
			topics = append(topics, topic)
		}
	}
	return topics
}

// IndexTopic 索引/更新单个主题。
func (s *searchService) IndexTopic(topic *model.Topic) {
	if !s.meiliEnabled() || topic == nil {
		return
	}
	docs := s.buildDocuments([]model.Topic{*topic})
	if len(docs) == 0 {
		return
	}
	if err := s.client.AddDocuments(s.index, docs); err != nil {
		logrus.Error("search: index topic failed, topicId=", topic.Id, ", err: ", err)
	}
}

// DeleteTopicIndex 删除主题索引。
// 新旧站编号可能重复，因此两个前缀都删一次（删除不存在的文档是无害的）。
func (s *searchService) DeleteTopicIndex(topicId int64) {
	if !s.meiliEnabled() {
		return
	}
	ids := []string{searchDocId(false, topicId), searchDocId(true, topicId)}
	if err := s.client.DeleteDocuments(s.index, ids); err != nil {
		logrus.Error("search: delete topic index failed, topicId=", topicId, ", err: ", err)
	}
}

// ReindexAll 全量重建新站索引。
func (s *searchService) ReindexAll() {
	if !s.meiliEnabled() {
		return
	}
	logrus.Info("search: start reindex all topics")
	processed := 0
	TopicService.Scan(func(topics []model.Topic) {
		docs := s.buildDocuments(topics)
		if len(docs) == 0 {
			return
		}
		if err := s.client.AddDocuments(s.index, docs); err != nil {
			logrus.Error("search: reindex batch failed: ", err)
			return
		}
		processed += len(docs)
	})
	logrus.Info("search: reindex all topics done, total=", processed)
}

// ReindexOldBBS 全量重建旧站索引。
func (s *searchService) ReindexOldBBS() {
	if !s.meiliEnabled() {
		return
	}
	logrus.Info("search: start reindex old bbs topics")
	processed := 0
	err := OldBBSService.ScanTopics(500, func(topics []model.Topic) error {
		docs := s.buildDocuments(topics)
		if len(docs) == 0 {
			return nil
		}
		if err := s.client.AddDocuments(s.index, docs); err != nil {
			return err
		}
		processed += len(docs)
		return nil
	})
	if err != nil {
		logrus.Error("search: reindex old bbs failed: ", err)
		return
	}
	logrus.Info("search: reindex old bbs done, total=", processed)
}

func (s *searchService) buildDocuments(topics []model.Topic) []meili.Document {
	if len(topics) == 0 {
		return nil
	}
	newTopicIds := make([]int64, 0, len(topics))
	for i := range topics {
		if !topics[i].IsOldBBS {
			newTopicIds = append(newTopicIds, topics[i].Id)
		}
	}
	tagNames := s.topicTagNames(newTopicIds)
	nodeNames := s.nodeNames()

	docs := make([]meili.Document, 0, len(topics))
	for i := range topics {
		topic := &topics[i]
		doc := meili.Document{
			ID:              searchDocId(topic.IsOldBBS, topic.Id),
			TopicId:         topic.Id,
			IsOldBBS:        topic.IsOldBBS,
			Title:           topic.Title,
			NodeId:          topic.NodeId,
			UserId:          topic.UserId,
			AccessLv:        topic.AccessLv,
			Status:          topic.Status,
			Recommend:       topic.Recommend,
			CreateTime:      topic.CreateTime,
			LastCommentTime: topic.LastCommentTime,
			ViewCount:       topic.ViewCount,
			CommentCount:    topic.CommentCount,
			LikeCount:       topic.LikeCount,
		}
		if topic.IsOldBBS {
			doc.AuthorName = topic.Author
			doc.NodeName = topic.Forum
			doc.Content = truncateRunes(pkghtml.GetHtmlText(topic.Content), searchContentMaxRunes)
		} else {
			if tags, ok := tagNames[topic.Id]; ok {
				doc.Tags = tags
			} else {
				doc.Tags = []string{}
			}
			doc.NodeName = nodeNames[topic.NodeId]
			if user := cache.UserCache.Get(topic.UserId); user != nil {
				doc.AuthorName = user.Nickname
			}
			doc.Content = truncateRunes(plainTextFromMarkdown(topic.Content), searchContentMaxRunes)
		}
		docs = append(docs, doc)
	}
	return docs
}

func (s *searchService) topicTagNames(topicIds []int64) map[int64][]string {
	result := make(map[int64][]string)
	if len(topicIds) == 0 {
		return result
	}
	topicTags := repositories.TopicTagRepository.Find(sqls.DB(), sqls.NewCnd().In("topic_id", topicIds))
	if len(topicTags) == 0 {
		return result
	}
	tagIds := make([]int64, 0, len(topicTags))
	for _, topicTag := range topicTags {
		tagIds = append(tagIds, topicTag.TagId)
	}
	tagNameMap := make(map[int64]string, len(tagIds))
	for _, tag := range repositories.TagRepository.Find(sqls.DB(), sqls.NewCnd().In("id", tagIds)) {
		tagNameMap[tag.Id] = tag.Name
	}
	for _, topicTag := range topicTags {
		if name, ok := tagNameMap[topicTag.TagId]; ok {
			result[topicTag.TopicId] = append(result[topicTag.TopicId], name)
		}
	}
	return result
}

func (s *searchService) nodeNames() map[int64]string {
	result := make(map[int64]string)
	for _, node := range repositories.TopicNodeRepository.Find(sqls.DB(), sqls.NewCnd().Eq("status", constants.StatusOk)) {
		result[node.Id] = node.Name
	}
	return result
}

func searchDocId(isOldBBS bool, topicId int64) string {
	if isOldBBS {
		return searchDocIdOldPrefix + strconv.FormatInt(topicId, 10)
	}
	return searchDocIdNewPrefix + strconv.FormatInt(topicId, 10)
}

func parseSearchDocId(docId string) (bool, int64, bool) {
	if strings.HasPrefix(docId, searchDocIdOldPrefix) {
		id, err := strconv.ParseInt(strings.TrimPrefix(docId, searchDocIdOldPrefix), 10, 64)
		if err != nil || id <= 0 {
			return true, 0, false
		}
		return true, id, true
	}
	if strings.HasPrefix(docId, searchDocIdNewPrefix) {
		id, err := strconv.ParseInt(strings.TrimPrefix(docId, searchDocIdNewPrefix), 10, 64)
		if err != nil || id <= 0 {
			return false, 0, false
		}
		return false, id, true
	}
	return false, 0, false
}

// escapeLike 转义 LIKE 通配符，避免用户输入 % / _ 变成通配符导致全表匹配。
func escapeLike(keyword string) string {
	keyword = strings.ReplaceAll(keyword, "\\", "\\\\")
	keyword = strings.ReplaceAll(keyword, "%", "\\%")
	keyword = strings.ReplaceAll(keyword, "_", "\\_")
	return keyword
}

func plainTextFromMarkdown(markdownStr string) string {
	if strings.TrimSpace(markdownStr) == "" {
		return ""
	}
	return pkghtml.GetHtmlText(pkgmarkdown.ToHTML(markdownStr))
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	count := 0
	for i := range s {
		if count == max {
			return s[:i]
		}
		count++
	}
	return s
}
