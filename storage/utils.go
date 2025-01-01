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
	"gorm.io/gorm"
)

func GetModelConfig(c *gin.Context) {
	modelType := c.Param("modelType")
	log.Printf("Getting configuration for %s", modelType)
	modelConfig := getModelConfig(modelType)
	c.JSON(http.StatusOK, modelConfig)
}

func GetRecords[R Model](c *gin.Context, records *[]R) {
	GetModelRecords(c, records, []string{})
}

//Callers don't have gin context
func GetAllRecords[R Model](records *[]R) {
	GetAllModelRecords(records, []string{})
}

func GetModelRecords[R Model](c *gin.Context, records *[]R, modelTypes []string) {
	query := c.DefaultQuery("query", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

    db, err := GetDb(c)
	if err != nil {
		return
	}
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
				db = db.Where(whereClause)
			} else if field["type"] == "date" {
				whereClause := fmt.Sprintf("%s.%s::date %s '%s'", tableName, fieldName, filterOperator, filterValue)
				if filterOperator == "between" {
					whereClause = fmt.Sprintf("%s::date AND '%s'", whereClause, filterValue2)
				}
				db = db.Where(whereClause)
			} else if field["type"] == "select" {
				selectorOf := field["selectorOf"].(string)
				selectorModel, _ := getModel(selectorOf)
				selectorTableName := callFunctionGeneric(selectorModel, "TableName")

				query, args := processFieldFilter(filterOperator, filterValue, selectorTableName, "name")
				db = db.Joins("left join "+selectorTableName+" on "+selectorTableName+".id = "+tableName+"."+fieldName).
					Where(query, args...)
			} else {
				query, args := processFieldFilter(filterOperator, filterValue, tableName, fieldName)
				db = db.Where(query, args...)
			}
		}
	}

	sortFields := strings.Split(c.DefaultQuery("sort", ""), ",")
	for _, field := range sortFields {
		db = db.Order(field)
	}

	count, currentPage, totalPages := getModelRecords(db, query, page, pageSize, records, modelTypes)
	for i := range *records {
		callFunction(&(*records)[i], "PostLoad")
	}
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

//Callers don't have gin context
func GetAllModelRecords[R Model](records *[]R, modelTypes []string) {
	getModelRecords(GetDbSpecial(), "", 1, 1000, records, modelTypes)
}

func getModelRecords[R Model](db *gorm.DB, query string, page int, pageSize int, records *[]R, modelTypes []string) (count int64, currentPage int, totalPages int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	for i := 0; i < len(modelTypes); i++ {
		db = db.Preload(modelTypes[i])
	}
	if query != "" {
		tableName := callFunctionSlice(records, "TableName")
		db = db.Where(tableName+".name ILIKE ?", "%"+query+"%")
	}
	if condition := callFunctionSlice(records, "PreFetchConditions"); condition != "" {
		db = db.Where(condition)
	}
	if sort := callFunctionSlice(records, "PreFetchSort"); sort != "" {
		db = db.Order(sort)
	}

	var nilRecord *R = nil
	db.Model(nilRecord).Count(&count)
	db.Offset(offset).Limit(pageSize).Find(records)

	currentPage = (offset / pageSize) + 1
	totalPages = int((count + int64(pageSize) - 1) / int64(pageSize))

	log.Printf("Found %d records", len(*records))
	return
}

func GetRecord[R Model](c *gin.Context, record *R) {
	id := c.Param("id")

    db, err := GetDb(c)
	if err != nil {
		return
	}
	if err := getRecordById(db, record, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found!"})
		return
	}
	c.JSON(http.StatusOK, record)
}

//Callers don't have gin context
func GetRecordById[R Model](record *R, id string) error {
	db := GetDbSpecial()
	return getRecordById(db, record, id)
}

func getRecordById[R Model](db *gorm.DB, record *R, id string) error {
	if id == "" {
		return fmt.Errorf("Can't get record with empty ID")
	}
	if cleanedId, _ := callFunction(record, "CleanId", reflect.ValueOf(id)); cleanedId != "" {
		id = cleanedId
	}
	if condition, _ := callFunction(record, "PreFetchConditions"); condition != "" {
		db = db.Where(condition)
	}
	callFunction(record, "PostLoad")
	return db.First(record, id).Error
}

func CreateRecord[R Model](c *gin.Context, record *R) {
	log.Println("Creating record from request")
	if err := c.ShouldBindJSON(record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("Loaded record from request")
    db, err := GetDb(c)
	if err != nil {
		return
	}
	if err := createModelRecord(db, record); err != nil {
		errorCode := http.StatusBadRequest
		if strings.HasPrefix(err.Error(), "conflict") {
			errorCode = http.StatusConflict
		}
		c.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	callFunction(record, "PostLoad")
	c.JSON(http.StatusOK, record)
}

//Callers don't have gin context
func CreateModelRecord[R Model](record *R) error {
    db := GetDbSpecial()
	return createModelRecord(db, record)
}

func createModelRecord[R Model](db *gorm.DB, record *R) error {
	if _, err := callFunction(record, "PreUpdate"); err != nil {
		return err
	}
	if err := db.Create(record).Error; err != nil {
		return err
	}
	log.Println("Record created successfully")
	return nil
}

func UpdateRecord[R Model](c *gin.Context, record *R) {
	id := c.Param("id")
    db, err := GetDb(c)
	if err != nil {
		return
	}
	if err := getRecordById(db, record, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found!"})
		return
	}
	if err := c.ShouldBindJSON(record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := persistRecord(db, record); err != nil {
		errorCode := http.StatusBadRequest
		if strings.HasPrefix(err.Error(), "conflict") {
			errorCode = http.StatusConflict
		}
		c.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	callFunction(record, "PostLoad")
	c.JSON(http.StatusOK, record)
}

//Callers don't have gin context
func PersistRecord[R Model](record *R) error {
    db := GetDbSpecial()
    return persistRecord(db, record);
}

func persistRecord[R Model](db *gorm.DB, record *R) error {
	if _, err := callFunction(record, "PreUpdate"); err != nil {
		return err
	}
	if err := db.Save(record).Error; err != nil {
		return err
	}
	log.Println("Record updated successfully")
	return nil
}

func DeleteRecord[R Model](c *gin.Context, record *R) {
	id := c.Param("id")
	if cleanedId, _ := callFunction(record, "CleanId", reflect.ValueOf(id)); cleanedId != "" {
		id = cleanedId
	}
    db, err := GetDb(c)
    if err != nil {
        return
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
	filterValue := c.DefaultQuery(fieldName, "")
	return strings.TrimSpace(filterValue)
}

func callFunctionGeneric(record interface{}, functionName string) string {
	modelType := reflect.TypeOf(record)
	output, _ := callFunctionType(modelType, record, functionName)
	return output
}

func callFunctionSlice[R Model](_ *[]R, functionName string) string {
	var nilRecord *R = nil
	output, _ := callFunction(nilRecord, functionName)
	return output
}

func callFunction[R Model](record *R, functionName string, extraParams ...reflect.Value) (outputStr string, outputErr error) {
	modelType := reflect.TypeOf(record)
	return callFunctionType(modelType, record, functionName, extraParams...)
}

func callFunctionType(modelType reflect.Type, record interface{}, functionName string, extraParams ...reflect.Value) (outputStr string, outputErr error) {
	if method, found := modelType.MethodByName(functionName); found {
		log.Printf("Found %s function, calling it", functionName)

		var params []reflect.Value
		params = append(params, reflect.ValueOf(record))
		params = append(params, extraParams...)

		results := method.Func.Call(params)
		if len(results) > 0 && results[0].Kind() == reflect.String {
			outputStr = results[0].String()
		} else if len(results) > 0 && !results[0].IsNil() && results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			outputErr = results[0].Interface().(error)
		}
	}
	return
}
