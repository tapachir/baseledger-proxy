package handler

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	uuid "github.com/kthomas/go.uuid"
	"github.com/unibrightio/proxy-api/dbutil"
	"github.com/unibrightio/proxy-api/logger"
	"github.com/unibrightio/proxy-api/restutil"
	"github.com/unibrightio/proxy-api/types"
)

type orgDetailsDto struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type createOrgRequest struct {
	Name string `json:"name"`
}

type deleteOrgRequest struct {
	Id string `json:"organization_id"`
}

func GetOrganizationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var organizations []types.Organization
		dbutil.Db.GetConn().Find(&organizations)

		var organizationDtos []orgDetailsDto

		for i := 0; i < len(organizations); i++ {
			organizationDto := &orgDetailsDto{}
			organizationDto.Id = organizations[i].Id
			organizationDto.Name = organizations[i].OrganizationName
			organizationDtos = append(organizationDtos, *organizationDto)
		}

		restutil.Render(organizationDtos, 200, c)
	}
}

func CreateOrganizationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, err := c.GetRawData()
		if err != nil {
			restutil.RenderError(err.Error(), 400, c)
			return
		}

		req := &createOrgRequest{}
		err = json.Unmarshal(buf, &req)
		if err != nil {
			restutil.RenderError(err.Error(), 422, c)
			return
		}

		newOrganization := newOrganization(*req)

		if !newOrganization.Create() {
			logger.Errorf("error when creating new organization")
			restutil.RenderError("error when creating new organization", 500, c)
			return
		}

		restutil.Render(newOrganization, 200, c)
	}
}

func DeleteOrganizationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgId := c.Param("id")

		var existingOrganization types.Organization
		dbError := dbutil.Db.GetConn().First(&existingOrganization, "id = ?", orgId).Error

		if dbError != nil {
			logger.Errorf("error trying to fetch organization with id %s\n", orgId)
			restutil.RenderError("organization not found", 404, c)
			return
		}

		if !existingOrganization.Delete() {
			logger.Errorf("error when deleting organization")
			restutil.RenderError("error when deleting organization", 500, c)
			return
		}

		restutil.Render(nil, 200, c)
	}
}

func newOrganization(req createOrgRequest) *types.Organization {
	return &types.Organization{
		OrganizationName: req.Name,
	}
}