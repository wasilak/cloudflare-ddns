package web

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/wasilak/cloudflare-ddns/libs/api"
	"github.com/wasilak/cloudflare-ddns/libs/cf"
)

func (s *Server) healthRoute(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

func (s *Server) apiList(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, api.Records)
}

func (s *Server) apiDelete(c echo.Context) error {
	recordName := c.Param("record_name")
	zoneName := c.Param("zone_name")

	var response map[string]any
	var httpStatus int

	_, err := api.DeleteRecord(c.Request().Context(), recordName, zoneName)
	if err != nil {
		if err.Error() == "record not found" {
			response = map[string]any{
				"message":    "Record not foud",
				"recordName": recordName,
				"zoneName":   zoneName,
				"error":      err.Error(),
			}
			httpStatus = http.StatusNotFound
		} else {
			response = map[string]any{
				"message":    "Record not deleted",
				"recordName": recordName,
				"zoneName":   zoneName,
				"error":      err.Error(),
			}
			httpStatus = http.StatusInternalServerError
		}
	} else {
		response = map[string]any{
			"message":    "Record deleted",
			"recordName": recordName,
			"zoneName":   zoneName,
		}
		httpStatus = http.StatusOK
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(httpStatus, response)
}

func (s *Server) apiCreate(c echo.Context) error {
	record := cf.ExtendedCloudflareDNSRecord{}
	if err := c.Bind(&record); err != nil {
		return err
	}

	var response map[string]any

	slog.InfoContext(c.Request().Context(), "Creating record", "record", record)

	_, err := api.AddRecord(c.Request().Context(), &record)
	if err != nil {
		response = map[string]any{
			"message": "Record not created",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	response = map[string]any{
		"message": "Record created",
		"record":  record,
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusCreated, response)
}

func (s *Server) apiUpdate(c echo.Context) error {
	record := cf.ExtendedCloudflareDNSRecord{}
	if err := c.Bind(&record); err != nil {
		return err
	}

	var response map[string]any

	_, err := api.UpdateRecord(c.Request().Context(), &record)
	if err != nil {
		response = map[string]any{
			"message": "Record not updated",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	response = map[string]any{
		"message": "Record updated",
		"record":  record,
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusCreated, response)
}
