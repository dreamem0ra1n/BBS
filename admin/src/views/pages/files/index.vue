<template>
  <section v-loading="listLoading" class="page-container">
    <div ref="toolbar" class="toolbar">
      <el-button
        type="danger"
        :loading="deleting"
        :disabled="busy || selectedRows.length === 0"
        @click="deleteSelected"
        >删除选中文件</el-button
      >
      <el-form :inline="true" :model="filters">
        <el-form-item><el-input v-model="filters.fileName" placeholder="文件名" /></el-form-item>
        <el-form-item><el-input v-model="filters.topicId" placeholder="帖子编号" /></el-form-item>
        <el-form-item><el-input v-model="filters.commentId" placeholder="评论编号" /></el-form-item>
        <el-form-item>
          <el-select v-model="filters.sourceType" clearable placeholder="来源" @change="list">
            <el-option label="帖子" value="topic" />
            <el-option label="评论" value="comment" />
            <el-option label="未记录来源" value="unattached" />
            <el-option label="用户头像" value="avatar" />
            <el-option label="用户背景图" value="background" />
            <el-option label="节点图标" value="node_logo" />
            <el-option label="友情链接图标" value="link_logo" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="list">查询</el-button></el-form-item>
      </el-form>
      <el-alert
        title="删除时自动检查实际引用，仅删除未被使用的文件。来源标签不代表当前是否仍在使用。"
        type="info"
        :closable="false"
      />
      <el-alert v-if="operationError" :title="operationError" type="error" :closable="false" />
    </div>

    <div ref="mainContent" :style="{ height: mainHeight }" class="page-section">
      <el-table :data="results" stripe height="100%" @selection-change="selectedRows = $event">
        <el-table-column type="selection" width="55" />
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
            <el-tag v-else type="info">{{ scope.row.sourceLabel || "未记录来源" }}</el-tag>
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
      deleting: false,
      operationError: "",
    };
  },
  computed: {
    busy() {
      return this.listLoading || this.deleting;
    },
  },
  mounted() {
    mainHeight(this);
    this.list();
  },
  methods: {
    async list() {
      if (this.busy) return;
      this.listLoading = true;
      this.selectedRows = [];
      const params = {
        file_name: this.filters.fileName,
        topic_id: this.filters.topicId,
        comment_id: this.filters.commentId,
        source_type: this.filters.sourceType,
        page: this.page.page,
        limit: this.page.limit,
      };
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
    async deleteSelected() {
      if (this.busy || this.selectedRows.length === 0) return;
      const rows = [...this.selectedRows];
      this.deleting = true;
      try {
        await this.$confirm(
          `检查并删除选中的 ${rows.length} 个文件？仍被引用的文件会跳过，无引用的文件将永久删除。`,
          "确认删除",
          { type: "warning" }
        );
      } catch (error) {
        this.deleting = false;
        return;
      }
      this.operationError = "";
      let deletedCount = 0;
      const failures = [];
      try {
        for (const file of rows) {
          try {
            await this.axios.form("/api/admin/file/delete", { id: file.id }, { timeout: 30000 });
            deletedCount += 1;
          } catch (error) {
            failures.push(`${file.file_name}：${error.message || "请求失败"}`);
          }
        }
        if (failures.length) {
          this.operationError = `已删除 ${deletedCount} 个文件，${
            failures.length
          } 个未删除。${failures.join("；")}`;
        } else {
          this.$message.success(`已删除 ${deletedCount} 个文件`);
        }
      } finally {
        this.deleting = false;
        this.selectedRows = [];
        await this.list();
      }
    },
  },
};
</script>
