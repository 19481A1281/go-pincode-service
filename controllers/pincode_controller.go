package controllers

import (
	"net/http"
	"strconv"

	"github.com/19481A1281/go-pincode-service/models"
	"github.com/19481A1281/go-pincode-service/services"
	"github.com/gin-gonic/gin"
)

type PincodeController struct {
	service services.PincodeService
}

func NewPincodeController(service services.PincodeService) *PincodeController{
	return &PincodeController{service: service}
}

func (c *PincodeController) Create(ctx *gin.Context){
	var pincode models.Pincode

	if err := ctx.ShouldBindJSON(&pincode); err != nil{
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error" : err.Error(),
		})
		return
	}

	pin, err := c.service.Create(&pincode)
	if err!=nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error" : err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, pin)
}

func (c *PincodeController) GetByID(ctx *gin.Context){

	id,_ := strconv.Atoi(ctx.Param("id"))

	pincode, err := c.service.GetByID(uint16(id))

	if err!=nil{
		ctx.JSON(http.StatusNotFound, gin.H{
			"error" : err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, pincode)
}

func (c *PincodeController) GetByPincode(ctx *gin.Context){
	pincode,_ := strconv.Atoi(ctx.Param("pincode"))

	data, err := c.service.GetByPincode(uint32(pincode))

	if err != nil{
		ctx.JSON(http.StatusNotFound,gin.H{
			"error" : err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (c *PincodeController) GetAll(ctx *gin.Context){

	page,_ := strconv.Atoi(ctx.DefaultQuery("page","1"))
	limit,_ := strconv.Atoi(ctx.DefaultQuery("limit","10"))

	pincodes, err := c.service.GetAll(page,limit)
	if err!=nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error" : err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, pincodes)
}

func(c *PincodeController) Update(ctx *gin.Context){
	pin,_ := strconv.Atoi(ctx.Param("pincode"))

	var updates map[string]interface{}

	if err := ctx.ShouldBindJSON(&updates); err!=nil{
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error" : err.Error(),
		})
		return
	}

	pincode, err := c.service.Update(uint32(pin),updates)

	if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error" : err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, pincode)
}

func(c *PincodeController) Delete(ctx *gin.Context){
	pin,_ := strconv.Atoi(ctx.Param("pincode"))
	
	err := c.service.Delete(uint16(pin))
	if err != nil{
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error" : err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message" : "pincode deleted successfully",
	})
}