# Description
This is a utils project to provide basic CURD operation for a GO project that has gorm and gin.
So the target of this module is transform model fields into configuration to load into your HTML page. These configurations include:
## API URL:
This will be used to execute API calls.
## Fields with properties
These properties include extras field that can help with the UI display: 
- "hidden"
- "optional"
- "block"
- "chartData"
- "tags"
- "short-span"
- "masterSelector"
- "href"
- "enum"
## Extra actions
These are extra customized actions per model

# Supported CURD functions
This module provides all the needed CURD operation, as an example:
## GetModelConfig
The initial method to retirieve the cashed configuration to the UI.
## GetRecords
This includes filtering and sorting from gin url, supports multiple operations per common data-types of the fields.
## GetModelRecords
Same like GetRecords. Moreover, it has the capability to add other models to load during the initial fetch, like dept-name in a users list.
## GetAllRecords
Same like GetRecords but without gin context, so doesn't have any filtering or sorting, but it loads the PreFetchConditions by reflection.
## GetAllModelRecords
Same like GetAllRecords without gin context, so doesn't have any filtering or sorting, but it loads the PreFetchConditions by reflection. Moreover, it has the capability to add other models to load during the initial fetch, like dept-name in a users list.
## GetRecord
Fetch a record by id from gin context, this also loads the PreFetchConditions by reflection.
## GetRecordById
Same like GetAllRecords without gin context.
## CreateRecord
Creates a record based on gin context. This methods executed the PreUpdate function on the model by reflection.
## CreateModelRecord
Creates the passed record without gin context. This methods executed the PreUpdate function on the model by reflection.
## UpdateRecord
Updates a record based on gin context. This methods executed the PreUpdate function on the model by reflection.
## PersistRecord
Updates the passed record without gin context. This methods executed the PreUpdate function on the model by reflection.
## DeleteRecord
Updates a record based on gin context.


# Supporting Model Reflection methods
These provide extra functionality to help with the display:

## GetTitle
The CURD page title to get used on the UI
## GetApiUrl
The API router as configured in your router
## ExtraActions
Extra actions to show up on the UI, and it must have an API configured in your router
## PreFetchConditions
Conditions to be apllied on the get all and getById functions
## PreFetchSort
Default sort while listing the output
## CleanId(id string)
In case the id has a prefix in some cases like "A1", this will be called to give the developer control to lean that up before getById, update, delete operations
## PreUpdate
This method is called before creating/updating a model to possible modify the fields before saving it to the db.

# The module that uses this modeuls should do the following:
## Call security.ConfigureJWT([]byteP{})
## Have a dashboard page to redrect to once login is successful
## Should define a nav_bar as a template


# Build
GOOS=linux GOARCH=amd64 go build -o build/go-creative-utils

# Deployment (After commiting the code-changes)
-> Search for v0.1.29 version in the utils and update it
-> Commit the code
git tag v0.1.29
git push origin v0.1.29
-> Then update the dependancies in the child projects, by searching for: utils@v0.1.29
-> Then download the new version: go get -u github.com/ahmedsaleh747/go-creative-utils@v0.1.29
