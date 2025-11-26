package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kratos-admin/app/admin/service/internal/data/gorm/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/timeutil"
	pagination "github.com/tx7do/kratos-bootstrap/api/gen/go/pagination/v1"
	"gorm.io/gorm"

	adminV1 "kratos-admin/api/gen/go/admin/service/v1"
)

type MenuRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewMenuRepo(db *gorm.DB, logger log.Logger) *MenuRepo {
	return &MenuRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "menu/repo/admin-service")),
	}
}

// Count 统计菜单数量
func (r *MenuRepo) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Menu{})

	// 应用查询条件
	if err := r.applyConditions(query, conditions); err != nil {
		return 0, err
	}

	if err := query.Count(&count).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

// List 获取菜单列表
func (r *MenuRepo) List(ctx context.Context, req *pagination.PagingRequest, treeTravel bool) (*adminV1.ListMenuResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var menus []models.Menu
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Menu{})

	// 构建查询条件
	conditions, err := r.buildConditions(req)
	if err != nil {
		return nil, err
	}

	// 应用查询条件
	if err := r.applyConditions(query, conditions); err != nil {
		return nil, err
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query count failed")
	}

	// 分页查询
	if !req.GetNoPaging() {
		offset := (req.GetPage() - 1) * req.GetPageSize()
		query = query.Offset(int(offset)).Limit(int(req.GetPageSize()))
	}

	// 排序
	if len(req.GetOrderBy()) > 0 {
		query = query.Order(req.GetOrderBy()[0])
	} else {
		query = query.Order("created_at DESC")
	}

	// 字段掩码处理
	if req.GetFieldMask() != nil && len(req.GetFieldMask().GetPaths()) > 0 {
		query = query.Select(req.GetFieldMask().GetPaths())
	}

	if err := query.Find(&menus).Error; err != nil {
		r.log.Errorf("query list failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query list failed")
	}

	// 转换为 DTO
	dtos := make([]*adminV1.Menu, 0, len(menus))
	if treeTravel {
		// 构建树形结构
		for _, menu := range menus {
			if menu.ParentID == nil || *menu.ParentID == 0 {
				dto := r.toDTO(&menu)
				dtos = append(dtos, dto)
			}
		}
		for _, menu := range menus {
			if menu.ParentID != nil && *menu.ParentID != 0 {
				dto := r.toDTO(&menu)

				if r.travelChild(dtos, dto) {
					continue
				}

				dtos = append(dtos, dto)
			}
		}
	} else {
		for _, menu := range menus {
			dto := r.toDTO(&menu)
			dtos = append(dtos, dto)
		}
	}

	return &adminV1.ListMenuResponse{
		Total: uint32(total),
		Items: dtos,
	}, nil
}

// Get 根据ID获取菜单
func (r *MenuRepo) Get(ctx context.Context, req *adminV1.GetMenuRequest) (*adminV1.Menu, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	var menu models.Menu
	if err := r.db.WithContext(ctx).First(&menu, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, adminV1.ErrorNotFound("menu not found")
		}
		r.log.Errorf("query menu failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("query data failed")
	}

	return r.toDTO(&menu), nil
}

// Create 创建菜单
func (r *MenuRepo) Create(ctx context.Context, req *adminV1.CreateMenuRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	menu := r.fromCreateRequest(req)

	// 设置创建时间
	if req.Data.CreatedAt == nil {
		now := time.Now()
		menu.CreatedAt = now
		menu.UpdatedAt = now
	} else {
		createTime := timeutil.TimestamppbToTime(req.Data.CreatedAt)
		menu.CreatedAt = *createTime
		menu.UpdatedAt = *createTime
	}

	if err := r.db.WithContext(ctx).Create(&menu).Error; err != nil {
		r.log.Errorf("create menu failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

// Update 更新菜单
func (r *MenuRepo) Update(ctx context.Context, req *adminV1.UpdateMenuRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetData().GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &adminV1.CreateMenuRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	// 处理字段掩码
	updateData := make(map[string]interface{})
	if req.UpdateMask != nil {
		req.UpdateMask.Normalize()
		if !req.UpdateMask.IsValid(req.Data) {
			r.log.Errorf("invalid field mask [%v]", req.UpdateMask)
			return adminV1.ErrorBadRequest("invalid field mask")
		}

		// 根据字段掩码构建更新数据
		updateData = r.buildUpdateData(req.Data, req.UpdateMask.GetPaths())
	} else {
		// 更新所有字段
		updateData = r.buildUpdateDataFromRequest(req.Data)
	}

	// 设置更新时间
	updateData["updated_at"] = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.Menu{}).Where("id = ?", req.Data.GetId()).Updates(updateData).Error; err != nil {
		r.log.Errorf("update menu failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("update data failed")
	}

	return nil
}

// Delete 删除菜单
func (r *MenuRepo) Delete(ctx context.Context, req *adminV1.DeleteMenuRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	// TODO: Need to handle child menu deletion like in the original ent implementation
	// This would require recursively finding and deleting child menus

	if err := r.db.WithContext(ctx).Delete(&models.Menu{}, req.GetId()).Error; err != nil {
		r.log.Errorf("delete menu failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// IsExist 检查菜单是否存在
func (r *MenuRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Menu{}).Where("id = ?", id).Count(&count).Error; err != nil {
		r.log.Errorf("check menu exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return count > 0, nil
}

// travelChild 递归遍历子菜单
func (r *MenuRepo) travelChild(nodes []*adminV1.Menu, node *adminV1.Menu) bool {
	if nodes == nil {
		return false
	}

	if node.ParentId == nil {
		nodes = append(nodes, node)
		return true
	}

	for _, n := range nodes {
		if node.ParentId == nil {
			continue
		}

		if n.GetId() == node.GetParentId() {
			n.Children = append(n.Children, node)
			return true
		} else {
			if r.travelChild(n.Children, node) {
				return true
			}
		}
	}
	return false
}

// 辅助方法

// buildConditions 构建查询条件
func (r *MenuRepo) buildConditions(req *pagination.PagingRequest) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})

	// 处理查询字符串
	if req.GetQuery() != "" {
		// 这里需要解析查询字符串，例如：name:admin,status:1
		// 简化实现，实际需要更复杂的解析逻辑
		queryParts := strings.Split(req.GetQuery(), ",")
		for _, part := range queryParts {
			if kv := strings.Split(part, ":"); len(kv) == 2 {
				conditions[kv[0]] = kv[1]
			}
		}
	}

	// 处理 OR 查询
	if req.GetOrQuery() != "" {
		// 实现 OR 查询逻辑
	}

	return conditions, nil
}

// applyConditions 应用查询条件
func (r *MenuRepo) applyConditions(query *gorm.DB, conditions map[string]interface{}) error {
	for key, value := range conditions {
		switch key {
		case "name":
			query = query.Where("name LIKE ?", "%"+fmt.Sprintf("%v", value)+"%")
		case "status":
			query = query.Where("status = ?", value)
		case "parent_id":
			query = query.Where("parent_id = ?", value)
		case "type":
			query = query.Where("type = ?", value)
		case "tenant_id":
			query = query.Where("tenant_id = ?", value)
		default:
			query = query.Where(key+" = ?", value)
		}
	}
	return nil
}

// buildUpdateData 根据字段掩码构建更新数据
func (r *MenuRepo) buildUpdateData(data *adminV1.Menu, paths []string) map[string]interface{} {
	updateData := make(map[string]interface{})

	for _, path := range paths {
		switch path {
		case "parent_id":
			if data.ParentId != nil {
				updateData["parent_id"] = *data.ParentId
			}
		case "type":
			if data.Type != nil {
				updateData["type"] = data.Type.String()
			}
		case "path":
			if data.Path != nil {
				updateData["path"] = *data.Path
			}
		case "redirect":
			if data.Redirect != nil {
				updateData["redirect"] = *data.Redirect
			}
		case "name":
			if data.Name != nil {
				updateData["name"] = *data.Name
			}
		case "component":
			if data.Component != nil {
				updateData["component"] = *data.Component
			}
		case "status":
			if data.Status != nil {
				updateData["status"] = int32(*data.Status)
			}
			// 添加更多字段处理...
		}
	}

	return updateData
}

// buildUpdateDataFromRequest 从请求构建完整的更新数据
func (r *MenuRepo) buildUpdateDataFromRequest(data *adminV1.Menu) map[string]interface{} {
	updateData := make(map[string]interface{})

	if data.ParentId != nil {
		updateData["parent_id"] = *data.ParentId
	}
	if data.Type != nil {
		updateData["type"] = data.Type.String()
	}
	if data.Path != nil {
		updateData["path"] = *data.Path
	}
	if data.Redirect != nil {
		updateData["redirect"] = *data.Redirect
	}
	if data.Name != nil {
		updateData["name"] = *data.Name
	}
	if data.Component != nil {
		updateData["component"] = *data.Component
	}
	if data.Status != nil {
		updateData["status"] = int32(*data.Status)
	}

	return updateData
}

// fromCreateRequest 从创建请求构建模型
func (r *MenuRepo) fromCreateRequest(req *adminV1.CreateMenuRequest) *models.Menu {
	menu := &models.Menu{}

	if req.Data.ParentId != nil {
		menu.ParentID = req.Data.ParentId
	}
	if req.Data.Type != nil {
		menuType := req.Data.Type.String()
		menu.Type = &menuType
	}
	if req.Data.Path != nil {
		menu.Path = req.Data.Path
	}
	if req.Data.Redirect != nil {
		menu.Redirect = req.Data.Redirect
	}
	if req.Data.Name != nil {
		menu.Name = req.Data.Name
	}
	if req.Data.Component != nil {
		menu.Component = req.Data.Component
	}
	if req.Data.Status != nil {
		status := int32(*req.Data.Status)
		menu.Status = &status
	}
	if req.Data.CreatedBy != nil {
		menu.CreatedBy = req.Data.CreatedBy
	}
	if req.Data.UpdatedBy != nil {
		menu.UpdatedBy = req.Data.UpdatedBy
	}

	return menu
}

// toDTO 将模型转换为 DTO
func (r *MenuRepo) toDTO(menu *models.Menu) *adminV1.Menu {
	dto := &adminV1.Menu{
		Id: &menu.ID,
	}

	if menu.ParentID != nil {
		dto.ParentId = menu.ParentID
	}
	if menu.Type != nil {
		dto.Type = r.stringToMenuType(*menu.Type)
	}
	if menu.Path != nil {
		dto.Path = menu.Path
	}
	if menu.Redirect != nil {
		dto.Redirect = menu.Redirect
	}
	if menu.Name != nil {
		dto.Name = menu.Name
	}
	if menu.Component != nil {
		dto.Component = menu.Component
	}
	if menu.Status != nil {
		status := adminV1.Menu_Status(*menu.Status)
		dto.Status = &status
	}
	if menu.CreatedBy != nil {
		dto.CreatedBy = menu.CreatedBy
	}
	if menu.UpdatedBy != nil {
		dto.UpdatedBy = menu.UpdatedBy
	}

	// 设置时间字段
	dto.CreatedAt = timeutil.TimeToTimestamppb(&menu.CreatedAt)
	dto.UpdatedAt = timeutil.TimeToTimestamppb(&menu.UpdatedAt)

	// 初始化Children字段
	dto.Children = make([]*adminV1.Menu, 0)

	return dto
}

// 枚举转换辅助方法
func (r *MenuRepo) stringToMenuType(s string) *adminV1.Menu_Type {
	switch s {
	case "DIR":
		menuType := adminV1.Menu_FOLDER
		return &menuType
	case "MENU":
		menuType := adminV1.Menu_MENU
		return &menuType
	case "BUTTON":
		menuType := adminV1.Menu_BUTTON
		return &menuType
	case "EMBEDDED":
		menuType := adminV1.Menu_EMBEDDED
		return &menuType
	case "LINK":
		menuType := adminV1.Menu_LINK
		return &menuType
	default:
		return nil
	}
}
