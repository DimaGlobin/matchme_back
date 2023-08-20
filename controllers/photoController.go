package controllers

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/DimaGlobin/matchme/initializers"
	"github.com/DimaGlobin/matchme/models"
	"github.com/DimaGlobin/matchme/utils"
	"github.com/cheggaaa/pb/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UploadPhoto(c *gin.Context) {

	user, _ := GetUserFromReq(c)
	var photo models.Photo

	photoCount := initializers.DB.Model(&user).Association("Photos").Count()
	if photoCount > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You can't upload more than 5 photos",
		})
		return
	}

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Unavailable method",
		})
		return
	}

	file, uploadFile, err := c.Request.FormFile("file1") // Get file from request
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Content-type should be multipart/form-data",
		})
		return
	}

	fileSize, err := strconv.Atoi(c.GetHeader("Content-Length"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing file size in header",
		})
		return
	}

	fmt.Println("-----\n")

	fmt.Println("File info: ")
	fmt.Println("File size: ", fileSize)
	fmt.Println("File name: ", uploadFile.Filename)

	fmt.Println("\n-----\n")

	if fileSize > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File size should be less than 20 MB",
		})
		return
	}

	uploadDir := "../temp-files"
	filePath := uploadDir + "/" + uploadFile.Filename

	hash, err := utils.Hash([]byte(uploadFile.Filename))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to calculate hash",
		})
		return
	}

	var result *gorm.DB

	flags := os.O_WRONLY | os.O_CREATE
	if _, err := os.Stat(filePath); err == nil {
		result = initializers.DB.Unscoped().Delete(&photo, "hash = ?", hash)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to remove file from database",
			})
			return
		}
		resArr := utils.RemoveNumberFromArray(user.PhotoHashes, photo.Hash)
		if user.PhotoHashes == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to find hash in array",
			})
			return
		}

		user.PhotoHashes = resArr
		result = initializers.DB.Save(&user)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to remove hash form user",
			})
			return
		}
	}

	f, err := os.OpenFile(filePath, flags, 0666)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error opening a file",
		})
		return
	}
	defer f.Close()

	bufWriter := bufio.NewWriter(f)

	limit := int64(fileSize)

	bar := pb.Full.Start64(limit)

	// create proxy reader
	barReader := bar.NewProxyReader(file)

	if _, err := io.Copy(bufWriter, barReader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error writing to a file",
		})
		return
	}
	bufWriter.Flush()

	bar.Finish()

	photo = models.Photo{UserID: user.ID, ImageName: uploadFile.Filename, Hash: hash}
	result = initializers.DB.Create(&photo)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error to create note in database",
		})
		return
	}

	user.PhotoHashes = append(user.PhotoHashes, int64(photo.Hash))
	result = initializers.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save info about photos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Photo uploaded successfully",
	})
}

func GetPhoto(c *gin.Context) {
	user, _ := GetUserFromReq(c)

	hashStr := c.Param("hash")

	hash, err := strconv.ParseUint(hashStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse hash",
		})
		return
	}

	var photo models.Photo

	result := initializers.DB.First(&photo, "hash = ?", hash)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find photo in database",
		})
		return
	}

	if photo.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No rights to get this photo",
		})
		return
	}

	filePath := "../temp-files/" + photo.ImageName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	c.File(filePath)
}

func DeletePhoto(c *gin.Context) {
	user, _ := GetUserFromReq(c)

	hashStr := c.Param("hash")

	hash, err := strconv.ParseUint(hashStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't parse hash",
		})
		return
	}

	var photo models.Photo

	result := initializers.DB.First(&photo, "hash = ?", hash)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find photo in database",
		})
		return
	}

	if photo.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No rights to get this photo",
		})
		return
	}

	filePath := "../temp-files/" + photo.ImageName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	err = os.Remove(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed remove file",
		})
		return
	}

	user.PhotoHashes = utils.RemoveNumberFromArray(user.PhotoHashes, photo.Hash)
	if user.PhotoHashes == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find hash in array",
		})
		return
	}

	result = initializers.DB.Unscoped().Delete(&photo)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove file from database",
		})
		return
	}

	result = initializers.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove hash form user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Photo deleted successfully",
	})
}
