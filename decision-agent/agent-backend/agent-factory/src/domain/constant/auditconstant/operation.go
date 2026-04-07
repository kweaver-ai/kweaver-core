package auditconstant

const (
	CREATE    = "create"    // 新建/新建/Create
	DELETE    = "delete"    // 删除/刪除/Delete
	UPDATE    = "update"    // 修改/修改/Update
	COPY      = "copy"      // 复制/複製/Copy
	PUBLISH   = "publish"   // 发布/發布/Publish
	UNPUBLISH = "unpublish" // 取消发布/取消發布/Unpublish
	// NOTE： 新增操作类型

	IMPORT         = "import"         // 导入/導入/Import
	EXPORT         = "export"         // 导出/導出/Export
	MODIFY_PUBLISH = "modify_publish" // 更新发布/更新發布/Modify and Publish
	COPY_PUBLISH   = "copy_publish"   // 复制并发布/複製並發布/Copy and Publish

)
