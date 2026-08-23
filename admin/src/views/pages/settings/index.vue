<template>
  <section v-loading="loading" class="page-container">
    <div class="page-section config-panel">
      <el-tabs value="commonConfigTab">
        <el-tab-pane label="通用配置" name="commonConfigTab">
          <el-form label-width="160px">
            <el-form-item label="网站名称">
              <el-input v-model="config.siteTitle" type="text" placeholder="网站名称" />
            </el-form-item>

            <el-form-item label="网站描述">
              <el-input
                v-model="config.siteDescription"
                type="textarea"
                autosize
                placeholder="网站描述"
              />
            </el-form-item>

            <el-form-item label="网站关键字">
              <el-select
                v-model="config.siteKeywords"
                style="width: 100%"
                multiple
                filterable
                allow-create
                default-first-option
                placeholder="网站关键字"
              />
            </el-form-item>

            <el-form-item label="网站公告">
              <el-input
                v-model="config.siteNotification"
                type="textarea"
                autosize
                placeholder="网站公告（支持输入HTML）"
              />
            </el-form-item>

            <el-form-item label="默认节点">
              <el-select
                v-model="config.defaultNodeId"
                style="width: 100%"
                placeholder="发帖默认节点"
              >
                <el-option
                  v-for="node in nodes"
                  :key="node.id"
                  :label="node.name"
                  :value="node.id"
                />
              </el-select>
            </el-form-item>
            <!--<template v-if="config.loginMethod">
              <el-form-item label="登录方式">
                <el-checkbox v-model="config.loginMethod.password"> 密码登录 </el-checkbox>
                <el-checkbox v-model="config.loginMethod.qq"> QQ登录 </el-checkbox>
                <el-checkbox v-model="config.loginMethod.github"> Github登录 </el-checkbox>
                <el-checkbox v-model="config.loginMethod.osc"> OSChina登录 </el-checkbox>
              </el-form-item>
            </template>-->
            <el-form-item label="站外链接跳转页面">
              <el-tooltip content="在跳转前需手动确认是否前往该站外链接" placement="top">
                <el-switch v-model="config.urlRedirect" />
              </el-tooltip>
            </el-form-item>

            <el-form-item label="启用评论可见内容">
              <el-tooltip content="发帖时支持设置评论后可见内容" placement="top">
                <el-switch v-model="config.enableHideContent" />
              </el-tooltip>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="导航配置" name="navConfigTab" class="nav-panel">
          <draggable v-model="config.siteNavs" draggable=".nav" handle=".nav-sort-btn" class="navs">
            <div v-for="(nav, index) in config.siteNavs" :key="index" class="nav">
              <el-row :gutter="20">
                <el-col :span="1">
                  <i class="iconfont icon-sort nav-sort-btn" />
                </el-col>
                <el-col :span="10">
                  <el-input v-model="nav.title" type="text" size="small" placeholder="标题" />
                </el-col>
                <el-col :span="11">
                  <el-input v-model="nav.url" type="text" size="small" placeholder="链接" />
                </el-col>
                <el-col :span="2">
                  <el-button
                    type="danger"
                    icon="el-icon-delete"
                    circle
                    size="small"
                    @click="delNav(index)"
                  />
                </el-col>
              </el-row>
            </div>
          </draggable>
          <div class="add-nav">
            <el-tooltip class="item" effect="dark" content="点击按钮添加导航" placement="top">
              <el-button type="primary" icon="el-icon-plus" circle @click="addNav" />
            </el-tooltip>
          </div>
        </el-tab-pane>
        <el-tab-pane v-if="config.scoreConfig" label="积分配置" name="scoreConfigTab">
          <el-form label-width="160px">
            <el-form-item label="发帖积分">
              <el-input-number
                v-model="config.scoreConfig.postTopicScore"
                :min="1"
                type="text"
                placeholder="发帖获得积分"
              />
            </el-form-item>
            <el-form-item label="跟帖积分">
              <el-input-number
                v-model="config.scoreConfig.postCommentScore"
                :min="1"
                type="text"
                placeholder="跟帖获得积分"
              />
            </el-form-item>
            <el-form-item label="签到积分上限">
              <el-input-number
                v-model="config.scoreConfig.checkInScoreMax"
                :min="1"
                type="text"
                placeholder="连续签到每日积分上限"
              />
            </el-form-item>
            <el-form-item label="单次赠米上限">
              <el-input-number
                v-model="config.scoreConfig.giftScoreMax"
                :min="1"
                :max="50"
                type="text"
                placeholder="单次赠米数量上限"
              />
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="生日随机祝福" name="birthdayBlessingTab">
          <el-form label-width="160px">
            <el-form-item label="开启生日随机祝福">
              <el-switch
                v-model="config.birthdayRandomBlessing"
                @change="onBirthdayBlessingChange"
              />
            </el-form-item>
          </el-form>
          <template v-if="config.birthdayRandomBlessing">
            <el-form :inline="true" :model="blessingFilters">
              <el-form-item
                ><el-input v-model="blessingFilters.nickname" placeholder="昵称"
              /></el-form-item>
              <el-form-item
                ><el-input v-model="blessingFilters.department" placeholder="部门"
              /></el-form-item>
              <el-form-item
                ><el-input v-model="blessingFilters.content" placeholder="内容"
              /></el-form-item>
              <el-form-item
                ><el-button type="primary" @click="loadBlessings">查询</el-button></el-form-item
              >
              <el-form-item
                ><el-button type="primary" @click="blessingImportVisible = true"
                  >导入</el-button
                ></el-form-item
              >
              <el-form-item
                ><el-button type="primary" @click="blessingBatchVisible = true"
                  >批量导入</el-button
                ></el-form-item
              >
              <el-form-item
                ><el-button
                  type="danger"
                  :disabled="!selectedBlessings.length"
                  @click="deleteBlessings"
                  >删除</el-button
                ></el-form-item
              >
            </el-form>
            <el-table
              :data="blessingResults"
              border
              stripe
              @selection-change="selectedBlessings = $event"
            >
              <el-table-column type="selection" width="55" />
              <el-table-column prop="nickname" label="昵称" width="160" />
              <el-table-column prop="department" label="部门" width="180" />
              <el-table-column prop="content" label="内容" />
            </el-table>
            <el-pagination
              :current-page="blessingPage.page"
              :page-size="blessingPage.limit"
              :total="blessingPage.total"
              layout="total, sizes, prev, pager, next, jumper"
              :page-sizes="[20, 50, 100]"
              @current-change="
                (page) => {
                  blessingPage.page = page;
                  loadBlessings();
                }
              "
              @size-change="
                (limit) => {
                  blessingPage.limit = limit;
                  loadBlessings();
                }
              "
            />
          </template>
        </el-tab-pane>
        <el-tab-pane label="反作弊配置" name="spamConfigTab">
          <el-form label-width="160px">
            <el-form-item label="发帖验证码">
              <el-tooltip content="发帖时是否开启验证码校验" placement="top">
                <el-switch v-model="config.topicCaptcha" />
              </el-tooltip>
            </el-form-item>

            <el-form-item label="邮箱验证后发帖">
              <el-tooltip content="需要验证邮箱后才能发帖" placement="top">
                <el-switch v-model="config.createTopicEmailVerified" />
              </el-tooltip>
            </el-form-item>

            <el-form-item label="邮箱验证后发表文章">
              <el-tooltip content="需要验证邮箱后才能发表文章" placement="top">
                <el-switch v-model="config.createArticleEmailVerified" />
              </el-tooltip>
            </el-form-item>

            <el-form-item label="邮箱验证后评论">
              <el-tooltip content="需要验证邮箱后才能发表评论" placement="top">
                <el-switch v-model="config.createCommentEmailVerified" />
              </el-tooltip>
            </el-form-item>

            <el-form-item label="发表文章审核">
              <el-tooltip content="发布文章后是否开启审核" placement="top">
                <el-switch v-model="config.articlePending" />
              </el-tooltip>
            </el-form-item>

            <el-form-item label="用户观察期(秒)">
              <el-tooltip
                content="观察期内用户无法发表话题、动态等内容，设置为 0 表示无观察期。"
                placement="top"
              >
                <el-input-number v-model="config.userObserveSeconds" :min="0" :max="720" />
              </el-tooltip>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div style="text-align: right">
        <el-button :loading="loading" type="primary" @click="save"> 保存配置 </el-button>
      </div>

      <el-dialog :visible.sync="blessingImportVisible" title="导入生日祝福">
        <el-form label-width="80px">
          <el-form-item label="昵称"
            ><el-input v-model="blessingImportForm.nickname"
          /></el-form-item>
          <el-form-item label="部门"
            ><el-input v-model="blessingImportForm.department"
          /></el-form-item>
          <el-form-item label="内容"
            ><el-input v-model="blessingImportForm.content" type="textarea"
          /></el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="blessingImportVisible = false">取消</el-button
          ><el-button type="primary" @click="importBlessing">提交</el-button>
        </div>
      </el-dialog>
      <el-dialog :visible.sync="blessingBatchVisible" title="批量导入生日祝福">
        <p>CSV 每行三列：昵称、部门、内容；支持内容中的逗号和引号。</p>
        <input type="file" accept=".csv,text/csv" @change="readBlessingCsv" />
        <el-input v-model="blessingBatchData" type="textarea" :rows="10" />
        <div slot="footer">
          <el-button @click="blessingBatchVisible = false">取消</el-button
          ><el-button type="primary" @click="batchImportBlessings">提交</el-button>
        </div>
      </el-dialog>
    </div>
  </section>
</template>

<script>
import draggable from "vuedraggable";

export default {
  components: { draggable },
  data() {
    return {
      config: {},
      loading: false,
      autocompleteTags: [],
      autocompleteTagLoading: false,
      nodes: [],
      blessingFilters: { nickname: "", department: "", content: "" },
      blessingResults: [],
      blessingPage: { page: 1, limit: 20, total: 0 },
      selectedBlessings: [],
      blessingImportVisible: false,
      blessingBatchVisible: false,
      blessingImportForm: { nickname: "", department: "", content: "" },
      blessingBatchData: "",
    };
  },
  mounted() {
    this.load();
  },
  methods: {
    async load() {
      this.loading = true;
      try {
        this.config = await this.axios.get("/api/admin/sys-config/all");
        if (this.config.scoreConfig && !this.config.scoreConfig.checkInScoreMax) {
          this.config.scoreConfig.checkInScoreMax = this.config.scoreConfig.checkInScore || 1;
        }
        if (this.config.scoreConfig && !this.config.scoreConfig.giftScoreMax) {
          this.config.scoreConfig.giftScoreMax = 50;
        }
        this.nodes = await this.axios.get("/api/admin/topic-node/nodes");
        if (this.config.birthdayRandomBlessing) this.loadBlessings();
      } catch (err) {
        this.$notify.error({ title: "错误", message: err.message });
      } finally {
        this.loading = false;
      }
    },
    async save() {
      this.loading = true;
      try {
        await this.axios.form("/api/admin/sys-config/save", {
          config: JSON.stringify(this.config),
        });
        this.$message({ message: "提交成功", type: "success" });
        this.load();
      } catch (err) {
        this.$notify.error({ title: "错误", message: err.message });
      } finally {
        this.loading = false;
      }
    },
    onBirthdayBlessingChange(enabled) {
      if (enabled) this.loadBlessings();
    },
    async loadBlessings() {
      if (!this.config.birthdayRandomBlessing) return;
      const data = await this.axios.form("/api/admin/birthday-blessing/list", {
        ...this.blessingFilters,
        page: this.blessingPage.page,
        limit: this.blessingPage.limit,
      });
      this.blessingResults = data.results;
      this.blessingPage = data.page;
    },
    async importBlessing() {
      await this.axios.form("/api/admin/birthday-blessing/import", this.blessingImportForm);
      this.blessingImportVisible = false;
      this.blessingImportForm = { nickname: "", department: "", content: "" };
      this.loadBlessings();
    },
    async batchImportBlessings() {
      await this.axios.form("/api/admin/birthday-blessing/batch-import", {
        data: this.blessingBatchData,
      });
      this.blessingBatchVisible = false;
      this.blessingBatchData = "";
      this.loadBlessings();
    },
    readBlessingCsv(event) {
      const file = event.target.files && event.target.files[0];
      if (!file) return;
      const reader = new FileReader();
      reader.onload = () => {
        this.blessingBatchData = reader.result;
      };
      reader.readAsText(file, "UTF-8");
      event.target.value = "";
    },
    async deleteBlessings() {
      await this.$confirm("确定删除选中的祝福吗？", "提示", { type: "warning" });
      await this.axios.form("/api/admin/birthday-blessing/delete", {
        ids: this.selectedBlessings.map((item) => item.id).join(","),
      });
      this.selectedBlessings = [];
      this.loadBlessings();
    },
    addNav() {
      if (!this.config.siteNavs) {
        this.config.siteNavs = [];
      }
      this.config.siteNavs.push({
        title: "",
        url: "",
      });
    },
    delNav(index) {
      if (!this.config.siteNavs) {
        return;
      }
      this.config.siteNavs.splice(index, 1);
    },
  },
};
</script>

<style scoped lang="scss">
.config-panel {
  // margin: 20px;
  padding: 10px;
}

.config {
  padding: 10px 0;
}

.nav-panel {
  .navs {
    border: 1px solid #ddd;
    border-radius: 5px;

    .nav {
      padding: 5px 5px;
      margin: 0;

      &:not(:last-child) {
        border-bottom: 1px solid #ddd;
      }

      .nav-sort-btn {
        font-size: 21px;
        font-weight: 700;
        cursor: pointer;
        float: right;
      }
    }
  }

  .add-nav {
    margin-top: 20px;
    text-align: center;
  }
}
</style>
