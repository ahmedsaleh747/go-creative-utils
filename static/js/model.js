// Mock configuration fetcher
let config;
let pageSize = 20;
let filters = {};
let sortFields = [];
let dependencyConfigs = [];

async function loadConfiguration() {
    const path = window.location.pathname;
    const segments = path.split('/');
    let modelType = null;
    if (segments.length >= 3 && segments[1] === "model") {
        modelType = segments[2];
    }
    const response = await getModelConfig(modelType);
    config = response.data;

    loadDependencies();
    generateTableHeader();
    prepareFilters();
    loadFilterFromUrl();
    fetchEntries(true);
}

function loadFilterFromUrl() {
    const params = new URLSearchParams(window.location.search);
    params.forEach((value, key) => {
        if (key.startsWith("filter.")) {
            const [fieldName, subKey] = key.replace('filter.', '').split('-');
            if(subKey === "operator") {
                $(`#${fieldName}-filterOp`).val(value);
                $(`#${fieldName}-filterOp`).selectpicker('refresh');
                $(`#${fieldName}-filterOp`).trigger('change')
            } else if(subKey === "value") {
                $(`#${fieldName}-filterVal`).val(value);
            } else if(subKey === "value2") {
                $(`#${fieldName}-filterVal2`).val(value);
            }
            addFilter(fieldName);
        }
    });
}

function generateTableHeader() {
    $('#title').text(config.title);

    const header = $('.modelTable #tableHeader');
    const tableHeaderRow = $('<tr></tr>')
    header.append(tableHeaderRow)
    const tableHeaderFilterRow = $('<tr></tr>').attr("id", "filter-row")
    header.append(tableHeaderFilterRow)

    config.fields.forEach(field => {
        const columnHeader = $('<th></th>')
            .attr('class', 'resizable')
            .data("fieldName", field.name)
            .text(field.label);
        columnHeader.append($('<i></i>').attr('class', "sort-icon"))
        if (field['short-span']) {
            columnHeader.css("width", "15%")
        }
        if (field.block) {
            columnHeader.css("width", "60%")
        }
        tableHeaderRow.append(columnHeader);

        const columnFilterHeader = $('<td></td>');
        tableHeaderFilterRow.append(columnFilterHeader)

        addFieldFilter(columnFilterHeader, field);
    });
    tableHeaderRow.append($('<th></th>').text('Actions').css("width", "20%"));
    tableHeaderFilterRow.append($('<th></th>'))
}

async function fetchEntries(clear) {
    const loadingFlag = $('#loading');
    if (clear) clearLoading();
    if (isLoading() || isLastPage()) return;

    const page = loadNextPage(loadingFlag);
    const response = await secureFetch(`${config.apiUrl}?page=${page}&pageSize=${pageSize}`, {
        headers: {'Content-Type': 'application/json'},
        data: {
            ...flattenFilters(),
            sort: sortFields,
        }
    });
    const data = response.data.items?response.data.items:[];
    loadingFlag.hide();
    const body = $('#tableBody');
    if (clear) body.empty();

    if (data.length === 0) loadingFlag.data("lastPage", true);
    else {
        $('#totalCount').html(response.data.total)
        $('#currentPage').html(response.data.currentPage)
        $('#totalPages').html(response.data.totalPages)
    }
    $('#serverTime').html(displayFormattedDate(response.data.serverTime))
    data.forEach(modelRecord => {
        const modelRow = $('<tr></tr>')
        body.append(modelRow);

        config.fields.forEach(field => {
            const fieldColumn = $('<td></td>');
            modelRow.append(fieldColumn);

            displayField(modelRecord, field, fieldColumn, false);
        });

        const modelActionsColumn = $('<td></td>')
        modelRow.append(modelActionsColumn);

        const editBtn = $('<button></button>')
            .attr('class', 'edit-btn')
            .data('id', modelRecord.id)
            .click(function (e) {
                const id = $(this).data('id');
                const modelRow = $(e.target).closest('tr');
                modelRow.empty();

                appendModelRow(modelRecord, modelRow, `${config.apiUrl}/${id}`, 'PUT');
            });
        editBtn.append($('<img>')
            .attr("src", `https://cdn.jsdelivr.net/gh/ahmedsaleh747/go-creative-utils@latest/static/images/edit.png`)
            .attr("alt", "Edit")
            .attr("style", "width: 24px; height: 24px")
        );
        const deleteBtn = $('<button></button>')
            .attr('class', 'delete-btn')
            .data('id', modelRecord.id)
            .click(async function (e) {
                if (confirm("Are you sure you want to delete this record?")) {
                    const id = $(this).data('id');
                    const response = await secureFetch(`${config.apiUrl}/${id}`, {method: 'DELETE'});
                    if (response.ok) fetchEntries(true);
                }
            });
        deleteBtn.append($('<img>')
            .attr("src", `https://cdn.jsdelivr.net/gh/ahmedsaleh747/go-creative-utils@latest/static/images/delete.png`)
            .attr("alt", "Delete")
            .attr("style", "width: 24px; height: 24px")
        )
        modelActionsColumn.append(editBtn);
        modelActionsColumn.append(deleteBtn);

        if (config.actions != null)
            config.actions.forEach(action => {
                const actionBtn = $('<button></button>')
                    .attr('class', 'edit-btn')
                    .data('id', modelRecord.id)
                    .click(function (e) {
                        const id = $(this).data('id');
                        return secureFetch(`${config.apiUrl}/${id}/${action}`)
                    });
                actionBtn.append($('<img>')
                    .attr("src", `/static/images/${action}.png`)
                    .attr("alt", action)
                    .attr("title", action)
                    .attr("style", "width: 24px; height: 24px")
                )
                modelActionsColumn.append(actionBtn);
            });
    });
}

function addRecordRow() {
    const modelTableBody = $('.modelTable #tableBody');
    const modelRow = $('<tr></tr>');
    modelTableBody.prepend(modelRow);

    appendModelRow(null, modelRow, config.apiUrl, 'POST');
}

function appendModelRow(modelRecord, modelRow, apiUrl, apiMethod) {
    const id = modelRecord == null ? '' : modelRecord['id'];
    config.fields.forEach(field => {
        appendFieldColumn(modelRecord, modelRow, field);
    });
    const modelActionsColumn = $('<td></td>')
    modelRow.append(modelActionsColumn);

    const saveBtn = $('<button></button>')
        .text("Save")
        .attr('class', 'save-btn btn btn-primary')
        .data('id', id)
        .click(async function () {
            const record = {};
            const inputs = modelRow.find('.input-field')
            inputs.each(function (_, input) {
                input = $(input)
                if (input.attr('name') == null || input.val() === '') return;
                if (input.attr('type') === 'password' && input.val() === '****') return;
                record[input.attr('name')] =
                    input.attr('type') === 'password'
                        ? CryptoJS.SHA256(input.val()).toString(CryptoJS.enc.Hex)
                        : input.val();
            });
            const response = await secureFetch(apiUrl, {
                method: apiMethod,
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(record)
            });
            if (response.ok) fetchEntries(true);
        });
    const cancelBtn = $('<button></button>')
        .text("Cancel")
        .attr('class', 'cancel-btn btn btn-secondary')
        .click(function () {
            if (modelRecord == null)
                modelRow.remove();
            else
                fetchEntries(true);   //Create new record
        })
    modelActionsColumn.append(saveBtn);
    modelActionsColumn.append(cancelBtn);
}

function appendFieldColumn(modelRecord, modelRow, field) {
    const fieldColumn = $('<td></td>');
    modelRow.append(fieldColumn);

    displayField(modelRecord, field, fieldColumn, true)
}

async function fetchDependencyData(field, query = '', page = 1, append = false, defaultOption) {
    if (field.loading || field.allLoaded) return;
    field.loading = true;
    field.lastPage = page;

    let masterSelectorValue = null
    if (field.masterSelector != null) {
        const masterFieldSelect = $('#' + field.masterSelector)
        masterSelectorValue = masterFieldSelect != null && masterFieldSelect.find(":selected").text()
        if (masterSelectorValue == null || masterSelectorValue === "") {
            console.log("Load master field " + field.masterSelector + " first")
            field.loading = false;
            return;
        }
    }

    //handle enums
    if (field.selectorOf === 'enum') {
        const data = field.allowedValues.map(value => {
            return {"id": value, "name": value}
        });
        processData(field, append, defaultOption, data);
        field.allLoaded = true;
        return;
    }

    let fieldFetchUrl = dependencyConfigs[field.selectorOf].apiUrl
    let masterFilter = "";
    if (masterSelectorValue != null) {
        masterFilter = `&${field.masterSelector}-operator=contains&${field.masterSelector}-value=${masterSelectorValue}`;
    }
    const response = await secureFetch(`${fieldFetchUrl}?page=${page}&query=${query}&pageSize=${pageSize}${masterFilter}&sort=name`, null, error => {
        console.error(`Error fetching ${field.name}:`, error);
        field.loading = false;
    });
    processData(field, append, defaultOption, response.data.items);
}

function processData(field, append, defaultOption, data) {
    const fieldSelect = $('.modelTable #' + field.name);
    if (!append) {
        fieldSelect.empty();
        if (defaultOption != null) fieldSelect.append(defaultOption);
    }
    if (data.length < pageSize) {
        field.allLoaded = true;
    }
    data
        .filter(model => defaultOption == null || model.id != defaultOption.value)
        .forEach(model => {
            fieldSelect.append(
                $('<option></option>')
                    .attr('value', model.id)
                    .attr('selected', false)
                    .text(model.name));
        });
    if (defaultOption == null) fieldSelect.val('default');
    fieldSelect.selectpicker('refresh');
    field.loading = false;
}

async function getModelConfig(modelType) {
    return await secureFetch(`/api/config/${modelType}`, {
        method: 'GET'
    });
}

async function loadDependencies() {
    const selectFields = config['fields']
        .filter(field => field.type === 'select')
        .filter(field => field.selectorOf !== 'enum');
    for (let i = 0; i < selectFields.length; i++) {
        const field = selectFields[i]
        const response = await getModelConfig(field.selectorOf);
        dependencyConfigs[field.selectorOf] = response.data;
    }
}

//---------------------------   TABLE PAGINATION START  ----------------------------------

function loadNextPage() {
    const loadingFlag = $('#loading')
    loadingFlag.show();
    const currentPage = loadingFlag.data("page");
    loadingFlag.data("page", currentPage + 1);
    return currentPage;
}

function clearLoading() {
    const loadingFlag = $('#loading')
    loadingFlag.data("page", 1)
    loadingFlag.removeData("lastPage")
}

function isLoading() {
    return $('#loading').css('display') !== 'none'
}

function isLastPage() {
    return $('#loading').data("lastPage")
}

//-----------------------   TABLE PAGINATION END  ------------------------------


//---------------------------   FIELDS START  ----------------------------------

function displayField(modelRecord, field, fieldColumn, editModel) {
    if (field.type === 'select') {
        displaySelectField(modelRecord, field, fieldColumn, editModel);
    } else if (field.chartData) {
        displayChartField(modelRecord, field, fieldColumn, editModel);
    } else if (field.block) {
        displayBlockField(modelRecord, field, fieldColumn, editModel);
    } else if (field.tags) {
        displayTagField(modelRecord, field, fieldColumn, editModel);
    } else {
        displayTextField(modelRecord, field, fieldColumn, editModel);
    }
}

function displaySelectField(modelRecord, field, fieldColumn, editMode) {
    const value = modelRecord == null ? '' : modelRecord[field.name];
    if (!editMode) {
        if (field.selectorOf !== 'enum') {
            if(value !== undefined && value != '') fieldColumn.text(modelRecord[field.selectorOf].name);
        } else {
            fieldColumn.text(value);
        }
        return;
    }

    const optionValue = value === undefined || value == '' || field.selectorOf == null ? null
        : field.selectorOf === 'enum'
            ? new Option(value, value, true, true)
            : new Option(modelRecord[field.selectorOf].name, value, true, true);

    field.allLoaded = false;
    field.loading = false;
    const fieldSelect = $('<select></select>')
        .attr('id', field.name)
        .attr('name', field.name)
        .attr('class', 'selectpicker input-field')
        .attr('data-live-search', field.selectorOf !== 'enum')
        .attr('data-none-results-text', 'No results matched {0}');
    fieldColumn.append(fieldSelect);

    if (optionValue != null) fieldSelect.append(optionValue);
    if (field.masterSelector != null) fieldSelect.attr("masterSelector", field.masterSelector);

    fieldSelect.on('shown.bs.select', function () {
        fetchDependencyData(field, '', 1, false, optionValue);
    });
    fieldSelect.on('changed.bs.select', function (e) {
        field.allLoaded = false
        fieldColumn.parent('tr').find(`[masterSelector=${field.name}]`).each(function (_, select) {
            select = $(select)
            select.empty();
            select.selectpicker('refresh');
            const field = config.fields.find(field => field.name === select.attr('name'));
            field.allLoaded = false
        });
    });
    fieldSelect.on('loaded.bs.select', function () {
        const searchBox = $(this).parent().find('.bs-searchbox input');
        searchBox.off('input').on('input', function () {
            const query = $(this).val();
            field.allLoaded = false;
            fetchDependencyData(field, query, 1, false, optionValue);
        });

        const dropdownMenu = fieldSelect.parent().find('.dropdown-menu .inner');
        dropdownMenu.off('scroll').on('scroll', function () {
            if (dropdownMenu.scrollTop() + dropdownMenu.innerHeight() >= dropdownMenu[0].scrollHeight) {
                fetchDependencyData(field, searchBox.val(), field.lastPage + 1, true);
            }
        });
    })

    fieldSelect.selectpicker('refresh');
}

function displayChartField(modelRecord, field, fieldColumn, editMode) {
    if (editMode) return
    const chartBtn = $('<button></button>')
        .text("Show Chart")
        .attr('class', 'save-btn btn btn-primary')
        .click(async function () {
            showChart(modelRecord['name'], modelRecord[field.name])
        });
    fieldColumn.append(chartBtn);
}

function displayTextField(modelRecord, field, fieldColumn, editMode) {
    const value = modelRecord == null ? '' : modelRecord[field.name];
    if (editMode && field.name != 'created_at' && field.name != 'updated_at') {
        const fieldInput = $('<input>')
            .attr('id', field.name)
            .attr('name', field.name)
            .attr('type', field.type)
            .attr('placeholder', field.label)
            .attr('class', 'input-field')
            .val(value);
        fieldColumn.append(fieldInput);
        return
    }
    if (field.href != null && modelRecord[field.href] != '') {
        const fieldLink = $('<a></a>')
            .attr("href", modelRecord[field.href])
            .attr("target", "_blank")
            .text(modelRecord[field.name]);
        fieldColumn.append(fieldLink);
    } else
        fieldColumn.text(displayFormattedValue(field, value));
}

function displayBlockField(modelRecord, field, fieldColumn, editMode) {
    if (editMode) {
        const value = modelRecord == null ? '' : modelRecord[field.name];
        let formattedValue = value.replace(/\n/g, "<br/>");
        const fieldInput = $('<textarea>')
            .attr('id', field.name)
            .attr('name', field.name)
            .attr('type', field.type)
            .attr('placeholder', field.label)
            .attr('class', 'input-field')
            .val(formattedValue);
        fieldColumn.append(fieldInput);
        return;
    }

    fieldColumn.addClass("long-text");
    if (modelRecord[field.name] === '') return

    const textValue = modelRecord[field.name]
    const formattedValue = textValue.replace(/\n/g, "<br/>");
    const isLongText = formattedValue.length > 120
    const shortText = isLongText? formattedValue.substring(0, 120) + "..." : formattedValue;
    const shortSpan = $('<span></span>')
        .html(shortText)
        .attr('class', 'short-text');
    fieldColumn.append(shortSpan);

    const longSpan = $('<span></span>')
        .html(formattedValue)
        .attr('class', 'full-text d-none');
    fieldColumn.append(longSpan);

    if(isLongText) {
        const toggleLink = $('<a></a>')
            .text("Read more")
            .attr('href', '#')
            .click(async function (event) {
                event.preventDefault(); // Prevent the default link behavior

                var $this = $(this);
                var $shortText = $this.siblings('.short-text');
                var $fullText = $this.siblings('.full-text');

                if ($fullText.hasClass('d-none')) {
                    $fullText.removeClass('d-none');
                    $shortText.addClass('d-none');
                    $this.html('Read less'); // Change the link text to "Read less"
                } else {
                    $fullText.addClass('d-none');
                    $shortText.removeClass('d-none');
                    $this.html('Read more'); // Change the link text back to "Read more"
                }
            });
        fieldColumn.append(toggleLink);
    }
}

function displayTagField(modelRecord, field, fieldColumn, editMode) {
    const value = modelRecord == null ? '' : modelRecord[field.name];
    const tagsArray = value.split(',');
    tagsArray.forEach(function (tag) {
        const tagSpan = $('<span></span>')
            .text(tag.trim())
            .attr('class', 'badge bg-info');
        fieldColumn.append(tagSpan);
    });
}

function showChart(title, chartData) {
    $('#chartTitle').text(title);

    const ctx = document.getElementById('exerciseChart').getContext('2d');
    const chartInstance = new Chart(ctx, {
        type: 'line',
        data: {
            labels: chartData.sensorData.map((_, i) => `Point ${i + 1}`),
            datasets: [{
                label: title,
                data: chartData.sensorData,
                borderColor: 'rgba(75, 192, 192, 1)',
                borderWidth: 2,
                fill: false
            }]
        },
        options: {
            responsive: true,
            scales: {
                x: {title: {display: true, text: 'Data Points'}},
                y: {title: {display: true, text: 'Value'}}
            }
        }
    });

    $('#chartModal').show();
    $('.close').on('click', function () {
        chartInstance.destroy();
        $('#chartModal').hide();
    });
}

function displayFormattedValue(field, value) {
    if(field.type != 'date' || value == '') return value;
    return displayFormattedDate(value);
}

function displayFormattedDate(value) {
    const date = new Date(value);
    const formattedDate = date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short', // e.g., "Sep"
        day: 'numeric'  // e.g., "7"
    });
    const formattedTime = date.toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit',
        hour12: true    // Use 12-hour format
    });

    // Create a two-line output in a span
    const displayDateTime = `${formattedDate}\n${formattedTime}`;
    return displayDateTime;
}

//---------------------------   FIELDS END  ----------------------------------


//---------------------   FIELD MANUPULATION START  --------------------------

function addFieldFilter(columnFilterHeader, field) {
    const fieldFilter = $('<div></div>').attr('class', 'input-group');
    columnFilterHeader.append(fieldFilter);

    const filterOp = filters[field.name] === undefined || filters[field.name].operator === '' ? '' : filters[field.name].operator;
    const filterVal = filters[field.name] === undefined || filters[field.name].value === '' ? '' : filters[field.name].value;
    const filterVal2 = filters[field.name] === undefined || filters[field.name].value2 === '' ? '' : filters[field.name].value2;
    const filterOpSelect = $('<select></select>')
        .attr('class', 'selectpicker custom-select')
        .attr('id', `${field.name}-filterOp`);
    fieldFilter.append(filterOpSelect);

    if (field.type === 'select' || field.type === 'text') {
        filterOpSelect.append(`
            <option value="blank">Blank</option>
            <option value="notBlank">Not Blank</option>
            <option value="equals">Equals</option>
            <option value="notEquals">Not Equals</option>
            <option value="contains">Contains</option>
            <option value="notContains">Not Contains</option>
            <option value="in">In</option>
        `);
    } else if (field.type === 'number' || field.type === 'date') {
        filterOpSelect.append(`
            <option value="=">=</option>
            <option value=">">></option>
            <option value=">=">>=</option>
            <option value="<"><</option>
            <option value="<="><=</option>
            <option value="between">Between</option>
        `);
    } else {
        filterOpSelect.append(`<option value="=">Equals</option>`);
    }
    filterOpSelect.val(filterOp);
    filterOpSelect.selectpicker();

    const fieldValInput = $('<input></input>')
        .attr('class', 'form-control')
        .attr('type', field.type === 'select' ? 'text' : field.type)
        .attr('id', `${field.name}-filterVal`)
        .attr('placeholder', `Filter by ${field.label}`)
        .data('fieldName', field.name)
        .css('display', 'none')
        .val(filterVal);
    fieldFilter.append(fieldValInput);
    if (field.type === 'number' || field.type === 'date') { //Handle the between op
        const fieldVal2Input = $('<input></input>')
            .attr('class', 'form-control')
            .attr('type', field.type)
            .attr('id', `${field.name}-filterVal2`)
            .attr('placeholder', `Filter by ${field.label}`)
            .data('fieldName', field.name)
            .css('display', 'none')
            .val(filterVal2);
        fieldFilter.append(fieldVal2Input);
    }
}

function prepareFilters() {
    if ($('.modelTable').data('filterInitialized')) return
    $('#filter-row select').on('change', function() {
        const $this = $(this)
        const fieldName = $this.attr('id').replace('-filterOp', '');
        if($this.val() === '' || $this.val() === null) {
            $(`#${fieldName}-filterVal`).hide();
            $(`#${fieldName}-filterVal2`).hide();
            return;
        }
        if ($this.val() === 'blank' || $this.val() === 'notBlank') {
            $(`#${fieldName}-filterVal`).hide();
            if($(`#${fieldName}-filterVal2`) !== null) $(`#${fieldName}-filterVal2`).hide();
        } else if ($this.val() === 'between') {
            $(`#${fieldName}-filterVal`).show();
            $(`#${fieldName}-filterVal2`).show();
        } else {
            $(`#${fieldName}-filterVal`).show();
            if($(`#${fieldName}-filterVal2`) !== null) $(`#${fieldName}-filterVal2`).hide();
        }
        if(addFilter(fieldName)) fetchEntries(true);
    });

    // Apply filters
    $('#filter-row input').on('focusout', function (e) {
        const $this = $(this)
        let fieldName = $this.data('fieldName');
        if(addFilter(fieldName)) fetchEntries(true);
    });

    // Remove filters
    $(document).on('click', '.remove-filter', function () {
        let fieldName = $(this).parent().attr('id').replace("badge-", "");
        deleteFilter(fieldName);
        $(`#${fieldName}-filterOp`).val('');
        $(`#${fieldName}-filterOp`).selectpicker('refresh');
        $(`#${fieldName}-filterOp`).trigger('change')
        $(`#${fieldName}-filterVal`).val('');
        if($(`#${fieldName}-filterVal2`) !== null) $(`#${fieldName}-filterVal2`).val('');
        fetchEntries(true);
    });

    // Sorting
    $('th.resizable').on('click', function () {
        const field = $(this).data('fieldName');
        let direction = $(this).hasClass('asc') ? 'desc' :
            $(this).hasClass('desc') ? null : 'asc';

        // Update the class for the header based on the direction
        $(this).removeClass('asc desc');
        sortFields = sortFields.filter(sort => !sort.startsWith(`${field} `)); // Remove existing sort for this field
        if (direction) {
            $(this).addClass(direction);
            $(this).find('.sort-icon').text(direction === 'asc' ? '↑' : '↓');
            sortFields.push(`${field} ${direction}`); // Add new sort if applicable
        } else {
            $(this).find('.sort-icon').text(''); // Clear the sort icon
        }
        fetchEntries(true);
    });

    $('.modelTable').data('filterInitialized', true)
}

function flattenFilters() {
    const flattened = {};

    Object.keys(filters).forEach(key => {
        const filter = filters[key];
        flattened[`${key}-operator`] = filter.operator;
        if (filter.value) {
            flattened[`${key}-value`] = filter.value;
        }
        if (filter.value2) {
            flattened[`${key}-value2`] = filter.value2;
        }
    });
    return flattened;
}

function addFilter(fieldName) {
    deleteFilter(fieldName);

    let filterOp = $(`#${fieldName}-filterOp`).val();
    let filterVal = $(`#${fieldName}-filterVal`).val();
    let filterVal2 = $(`#${fieldName}-filterVal2`) !== null? $(`#${fieldName}-filterVal2`).val() : '';

    if(filterOp === 'blank' || filterOp === 'notBlank') {
        filterVal = '';
        filterVal2 = '';
    } else if(
        (filterOp === 'between' && (filterVal === '' || filterVal === undefined || filterVal2 === '' || filterVal2 === undefined)) ||
        (filterOp !== 'between' && (filterVal === '' || filterVal === undefined))) {
        return false;     //missing filter value
    }

    filters[fieldName] = {
        "operator": filterOp
    };
    if(filterVal !== '' && filterVal !== undefined) filters[fieldName]["value"] = filterVal
    if(filterVal2 !== '' && filterVal2 !== undefined) filters[fieldName]["value2"] = filterVal2
    console.log(filters)

    const params = new URLSearchParams(window.location.search);
    Object.keys(filters).forEach(key => {
        params.set(`filter.${key}-operator`, filters[key].operator);
        if(filters[key].value !== '' && filters[key].value !== undefined) params.set(`filter.${key}-value`, filters[key].value);
        if(filters[key].value2 !== '' && filters[key].value2 !== undefined) params.set(`filter.${key}-value2`, filters[key].value2);
    });
    const newUrl = `${window.location.pathname}?${params.toString()}`;
    window.history.replaceState({}, '', newUrl);

    var filterOperator = $("option:selected", $(`#${fieldName}-filterOp`)).text();
    let filterStr = config.fields.find(field => field.name === fieldName).label;
    if(filterOp === 'blank' || filterOp === 'notBlank') filterStr += " is " + filterOperator;
    else if(filterOp === 'between') filterStr += " " + filterOperator + " " + filterVal + " and " + filterVal2;
    else filterStr += " " + filterOperator + " " + filterVal

    $('#filter-badges').append(`
        <span class="badge bg-primary filter-badge" id="badge-${fieldName}">${filterStr}
            <button type="button" class="btn-close remove-filter" aria-label="Close"></button>
        </span>
    `);
    return true
}

function deleteFilter(fieldName) {
    $(`#badge-${fieldName}`).remove();
    delete filters[fieldName];

    const params = new URLSearchParams(window.location.search);
    params.delete(`filter.${fieldName}-operator`);
    params.delete(`filter.${fieldName}-value`);
    params.delete(`filter.${fieldName}-value2`);

    const newUrl = `${window.location.pathname}?${params.toString()}`;
    window.history.replaceState({}, '', newUrl);
}

//---------------------   FIELD MANUPULATION END  ----------------------------


$(document).ready(function () {
    $(window).scroll(function () {
        if ($(window).scrollTop() + $(window).height() >= $(document).height() - 100 && !isLoading()) {
            fetchEntries(false);
        }
    });

    $('#addRecordBtn').on('click', addRecordRow);

    loadConfiguration();
});
