package storage

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var err error

//goland:noinspection SpellCheckingInspection
func InitDatabaseModels(dsn string, models []interface{}) {
	log.Printf("Configuring db connection for %d models ...", len(models))
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect to database")
	}

	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("failed to migrate database: %v\n", err)
		return
	}

	models = append(models, &User{})
	models = append(models, &Subscription{})
	for _, model := range models {
		AddConfig(model)
	}
}

func GetDb() *gorm.DB {
	return db
}

func GetModelConfig(c *gin.Context) {
	modelType := c.Param("modelType")
	log.Printf("Getting configuration for %s", modelType)
	modelConfig := getModelConfig(modelType)
	c.JSON(http.StatusOK, modelConfig)
}

func GetRecords[R Model](c *gin.Context, records *[]R) {
	GetModelRecords(c, records, []string{})
}

func GetAllRecords[R Model](records *[]R) {
	GetAllModelRecords(records, []string{})
}

func GetModelRecords[R Model](c *gin.Context, records *[]R, modelTypes []string) {
	query := c.DefaultQuery("query", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	tempDb := db
	recordType := reflect.TypeOf(records).Elem().Elem().Name()
	config := *getModelConfig(recordType)
	tableName := callFunctionSlice(records, "TableName")
	fields := config["fields"].([]map[string]any)
	for _, field := range fields {
		fieldName := field["name"].(string)
		if filterOperator := getFilterValue(c, fieldName+"-operator"); len(filterOperator) > 0 {
			filterValue := getFilterValue(c, fieldName+"-value")
			filterValue2 := getFilterValue(c, fieldName+"-value2")

			if field["type"] == "number" || field["type"] == "bool" {
				whereClause := fmt.Sprintf("%s.%s %s %s", tableName, fieldName, filterOperator, filterValue)
				if filterOperator == "between" {
					whereClause = fmt.Sprintf("%s AND %s", whereClause, filterValue2)
				}
				tempDb = tempDb.Where(whereClause)
			} else if field["type"] == "date" {
				whereClause := fmt.Sprintf("%s.%s::date %s '%s'", tableName, fieldName, filterOperator, filterValue)
				if filterOperator == "between" {
					whereClause = fmt.Sprintf("%s::date AND '%s'", whereClause, filterValue2)
				}
				tempDb = tempDb.Where(whereClause)
			} else if field["type"] == "select" {
				selectorOf := field["selectorOf"].(string)
				selectorModel, _ := getModel(selectorOf)
				selectorTableName := callFunctionGeneric(selectorModel, "TableName")

				query, args := processFieldFilter(filterOperator, filterValue, selectorTableName, "name")
				tempDb = tempDb.Joins("left join "+selectorTableName+" on "+selectorTableName+".id = "+tableName+"."+fieldName).
					Where(query, args...)
			} else {
				query, args := processFieldFilter(filterOperator, filterValue, tableName, fieldName)
				tempDb = tempDb.Where(query, args...)
			}
		}
	}

	sortFields := strings.Split(c.DefaultQuery("sort", ""), ",")
	for _, field := range sortFields {
		tempDb = tempDb.Order(field)
	}

	count, currentPage, totalPages := getModelRecords(tempDb, query, page, pageSize, records, modelTypes)
	c.JSON(http.StatusOK, gin.H{
		"total":       count,
		"currentPage": currentPage,
		"totalPages":  totalPages,
		"items":       records,
		"serverTime":  time.Now().Format(time.RFC3339),
	})
}

func processFieldFilter(filterOperator string, filterValue string, tableName string, fieldName string) (query interface{}, args []interface{}) {
	if filterOperator == "in" {
		var ilikeValues []string
		for _, val := range strings.Split(filterValue, ",") {
			ilikeValues = append(ilikeValues, fmt.Sprintf("'%%%s%%'", strings.TrimSpace(val)))
		}
		ilikeArray := strings.Join(ilikeValues, ",")
		whereClause := fmt.Sprintf("%s.%s ILIKE ANY (ARRAY[%s])", tableName, fieldName, ilikeArray)
		query = whereClause
	} else if filterOperator == "contains" {
		query = fmt.Sprintf("%s.%s ILIKE ?", tableName, fieldName)
		args = []interface{}{"%" + filterValue + "%"}
	} else if filterOperator == "notContains" {
		query = fmt.Sprintf("NOT %s.%s ILIKE ?", tableName, fieldName)
		args = []interface{}{"%" + filterValue + "%"}
	} else if filterOperator == "equals" {
		query = fmt.Sprintf("%s.%s = ?", tableName, fieldName)
		args = []interface{}{filterValue}
	} else if filterOperator == "notEquals" {
		query = fmt.Sprintf("NOT %s.%s = ?", tableName, fieldName)
		args = []interface{}{filterValue}
	} else if filterOperator == "blank" {
		query = fmt.Sprintf("%s.%s IS NULL OR %s.%s = ''", tableName, fieldName, tableName, fieldName)
	} else if filterOperator == "notBlank" {
		query = fmt.Sprintf("%s.%s IS NOT NULL AND NOT %s.%s = ''", tableName, fieldName, tableName, fieldName)
	}
	return
}

func GetAllModelRecords[R Model](records *[]R, modelTypes []string) {
	getModelRecords(db, "", 1, 1000, records, modelTypes)
}

func getModelRecords[R Model](db *gorm.DB, query string, page int, pageSize int, records *[]R, modelTypes []string) (count int64, currentPage int, totalPages int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	tempDb := db
	for i := 0; i < len(modelTypes); i++ {
		tempDb = tempDb.Preload(modelTypes[i])
	}
	if query != "" {
		tableName := callFunctionSlice(records, "TableName")
		tempDb = tempDb.Where(tableName+".name ILIKE ?", "%"+query+"%")
	}
	if condition := callFunctionSlice(records, "PreFetchConditions"); condition != "" {
		tempDb = tempDb.Where(condition)
	}
	if sort := callFunctionSlice(records, "PreFetchSort"); sort != "" {
		tempDb = tempDb.Order(sort)
	}

	var nilRecord *R = nil
	tempDb.Model(nilRecord).Count(&count)
	tempDb.Offset(offset).Limit(pageSize).Find(records)

	currentPage = (offset / pageSize) + 1
	totalPages = int((count + int64(pageSize) - 1) / int64(pageSize))

	log.Printf("Found %d records", len(*records))
	return
}

func GetRecord[R Model](c *gin.Context, record *R) {
	id := c.Param("id")

	if err := GetRecordById(record, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found!"})
		return
	}
	c.JSON(http.StatusOK, record)
}

func GetRecordById[R Model](record *R, id string) error {
	if cleanedId := callFunction(record, "CleanId", reflect.ValueOf(id)); cleanedId != "" {
		id = cleanedId
	}
	if condition := callFunction(record, "PreFetchConditions"); condition != "" {
		db = db.Where(condition)
	}
	return db.First(record, id).Error
}

func CreateRecord[R Model](c *gin.Context, record *R) {
	log.Println("Creating record from request")
	if err := c.ShouldBindJSON(record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("Loaded record from request")
	if err := CreateModelRecord(record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func CreateModelRecord[R Model](record *R) error {
	callFunction(record, "PreUpdate")
	if err := db.Create(record).Error; err != nil {
		return err
	}
	log.Println("Record created successfully")
	return nil
}

func UpdateRecord[R Model](c *gin.Context, record *R) {
	id := c.Param("id")
	if err := GetRecordById(record, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found!"})
		return
	}
	if err := c.ShouldBindJSON(record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := PersistRecord(record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func PersistRecord[R Model](record *R) error {
	callFunction(record, "PreUpdate")
	if err := db.Save(record).Error; err != nil {
		return err
	}
	log.Println("Record updated successfully")
	return nil
}

func DeleteRecord[R Model](c *gin.Context, record *R) {
	id := c.Param("id")
	if cleanedId := callFunction(record, "CleanId", reflect.ValueOf(id)); cleanedId != "" {
		id = cleanedId
	}
	if err := db.Delete(record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"action":  "Toast",
		"message": "Record deleted",
	})
}

func getFilterValue(c *gin.Context, fieldName string) string {
	return c.DefaultQuery(fieldName, "")
}

func callFunctionGeneric(record interface{}, functionName string) string {
	modelType := reflect.TypeOf(record)
	return callFunctionType(modelType, record, functionName)
}

func callFunctionSlice[R Model](_ *[]R, functionName string) string {
	var nilRecord *R = nil
	return callFunction(nilRecord, functionName)
}

func callFunction[R Model](record *R, functionName string, extraParams ...reflect.Value) string {
	modelType := reflect.TypeOf(record)
	return callFunctionType(modelType, record, functionName, extraParams...)
}

func callFunctionType(modelType reflect.Type, record interface{}, functionName string, extraParams ...reflect.Value) string {
	if method, found := modelType.MethodByName(functionName); found {
		log.Printf("Found %s function, calling it", functionName)

		var params []reflect.Value
		params = append(params, reflect.ValueOf(record))
		params = append(params, extraParams...)

		results := method.Func.Call(params)
		if len(results) > 0 && results[0].Kind() == reflect.String {
			return results[0].String()
		}
	}
	return ""
}
