<template>
  <section v-loading="listLoading" class="page-container">
    <div ref="toolbar" class="toolbar">
      <el-button type="danger" :disabled="selectedRows.length === 0" @click="deleteSelected">删除选中文件</el-button>
      <el-form :inline="true" :model="filters">
        <el-form-item><el-input v-model="filters.fileName" placeholder="文件名" /></el-form-item>
        <el-form-item><el-input v-model="filters.topicId" placeholder="帖子编号" /></el-form-item>
        <el-form-item><el-input v-model="filters.commentId" placeholder="评论编号" /></el-form-item>
        <el-form-item>
          <el-select v-model="filters.sourceType" clearable placeholder="来源" @change="list">
            <el-option label="帖子" value="topic" />
            <el-option label="评论" value="comment" />
            <el-option label="未关联" value="unattached" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="list">查询</el-button></el-form-item>
      </el-form>
    </div>

    <div ref="mainContent" :style="{ height: mainHeight }" class="page-section">
      <el-table :data="results" stripe height="100%" @selection-change="selectedRows = $event">
        <el-table-column type="selection" width="55" :selectable="isDeletable" />
        <el-table-column label="预览" width="100">
          <template slot-scope="scope">
            <el-image
              v-if="scope.row.content_type && scope.row.content_type.indexOf('image/') === 0"
              :src="scope.row.previewUrl"
              :preview-src-list="[scope.row.previewUrl]"
              fit="cover"
              style="width: 64px; height: 64px"
            />
            <el-link v-else :href="scope.row.previewUrl" target="_blank">打开</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
        <el-table-column prop="content_type" label="类型" width="150" />
        <el-table-column prop="file_size" label="大小" width="100" />
        <el-table-column label="来源" min-width="220">
          <template slot-scope="scope">
            <el-link v-if="scope.row.sourceUrl" :href="scope.row.sourceUrl" target="_blank">
              {{ scope.row.sourceLabel }}
            </el-link>
            <el-tag v-else type="info">未关联</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="上传时间" width="160">
          <template slot-scope="scope">{{ scope.row.create_time | formatDate }}</template>
        </el-table-column>
      </el-table>
    </div>

    <div ref="pagebar" class="pagebar">
      <el-pagination
        :page-sizes="[20, 50, 100, 300]"
        :current-page="page.page"
        :page-size="page.limit"
        :total="page.total"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handleLimitChange"
      />
    </div>
  </section>
</template>

<script>
import mainHeight from "@/utils/mainHeight";

export default {
  name: "Files",
  data() {
    return {
      mainHeight: "300px",
      results: [],
      listLoading: false,
      page: {},
      filters: {},
      selectedRows: [],
    };
  },
  mounted() {
    mainHeight(this);
    this.list();
  },
  methods: {
    async list() {
      this.listLoading = true;
      const params = Object.assign({}, this.filters, {
        page: this.page.page,
        limit: this.page.limit,
      });
      try {
        const data = await this.axios.form("/api/admin/file/list", params);
        this.results = data.results;
        this.page = data.page;
      } finally {
        this.listLoading = false;
      }
    },
    handlePageChange(value) {
      this.page.page = value;
      this.list();
    },
    handleLimitChange(value) {
      this.page.limit = value;
      this.list();
    },
    isDeletable(row) {
      return row.managed === true && row.source_type === "unattached" && !row.topic_id && !row.comment_id;
    },
    async deleteSelected() {
      await this.$confirm("删除选中的未关联文件？", "提示", { type: "warning" });
      for (const file of this.selectedRows) {
        await this.axios.form("/api/admin/file/delete", { id: file.id });
      }
      this.$message.success("删除成功");
      this.selectedRows = [];
      this.list();
    },
  },
};
</script>
