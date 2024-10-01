async function secureFetch(url, request, errorHandler) {
    const token = localStorage.getItem('token');
    if (!token) {
        redirectToLoginPage();
        return;
    }

    if (request != null && request.data != null) {
        const params = buildQueryParams(request.data);
        url = url.indexOf("?") === -1 ? `${url}?${params}` : `${url}&${params}`;
    }

    if (request == null) {
        request = {method: 'GET', headers: {'Authorization': 'Bearer ' + token}};
    } else if (request.headers == null) {
        request.headers = {'Authorization': 'Bearer ' + token};
    } else {
        request.headers['Authorization'] = 'Bearer ' + token;
    }

    const response = await fetch(url, request).catch(errorHandler);
    if (response?.status === 401 || response?.status === 403) {
        localStorage.removeItem('token');
        redirectToLoginPage();
        return
    } else if (response?.status >= 500 && response?.status <= 599) {
        showToast('error', 'Error!', 'Something went wrong! ' + response?.message, 10000);
    }
    // showToast('info', 'Information', 'Here is some important info.', 20000);

    let data;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
        data = await response.json();
        if (data != null && data.action != undefined) {
            if (executeAction(data) ) {
                return
            }
        } else if (data != null && data.actions != undefined) {
            data.actions.forEach(action => {
                executeAction(action);
            });
        }
    } else {
        data = await response.text();
        if(data === '') showToast('success', 'Success!', 'Request sent successfully!', 5000);
    }

    return {
        "ok": response.ok,
        "data": data
    };
}

function redirectToLoginPage() {
    // Clear the current UI using jQuery
    const body = $('body');
    body.empty();

    // Create and style the message container using jQuery and Bootstrap classes
    const messageContainer = $('<div></div>')
        .attr('class', 'redirect-message alert alert-warning text-center mt-5')
        .css('max-width', '600px')
        .css('margin', '0 auto');
    messageContainer.append($('<h4></h4>').text("Unauthorized"))
    messageContainer.append($('<p></p>').text("You will be redirected to the login page in 3 seconds..."));
    messageContainer.append($('<p></p>').html('If you aren\'t redirected, please press this <a href="/">link</a>.'));

    // Append the message container to the body
    body.append(messageContainer);

    setTimeout(() => {
        window.location.href = `/`;
    }, 3000);
}

function buildQueryParams(request) {
    return Object.keys(request)
        .map(key => encodeURIComponent(key) + '=' + encodeURIComponent(request[key]))
        .join('&');
}

function showToast(type, title, message, delay = 3000) {
    // Create the toast HTML structure
    const toastId = `toast-${Date.now()}`;
    const toastHtml = `
        <div id="${toastId}" class="toast toast-${type}" role="alert" aria-live="assertive" aria-atomic="true" data-bs-delay="${delay}">
            <div class="toast-progress"></div>
            <div class="toast-header">
                <strong class="mr-auto">${title}</strong>
                <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close"></button>
            </div>
            <div class="toast-body">${message}</div>
        </div>
    `;

    // Append the toast to the toast container
    $('.toast-container').append(toastHtml);

    // Initialize the toast with Bootstrap's toast functionality
    const $toast = $(`#${toastId}`);
    $toast.toast('show');

    const progressBar = $toast.find('.toast-progress');
    progressBar.css('width', '100%');
    progressBar.css('transition', `width ${delay}ms linear`);

    // Shrink the progress bar after a slight delay
    setTimeout(() => {
        progressBar.css('width', '0%');
    }, 100); // Start shrinking after a short delay

    // Remove the toast from the DOM after it hides
    $toast.on('hidden.bs.toast', function () {
        $(this).remove();
    });
}

/**
 * @param {The action details} data 
 * @returns true to stop further execution
 */
function executeAction(data) {
    if(data.action == 'Refresh') {
        location.reload()
        return true;
    } else if(data.action == 'Redirect') {
        window.open(data.url, '_blank');
        return true;
    } else if(data.action == 'Dialog') {
        showDialog(data)
    } else if(data.action == 'Toast') {
        showToast('success', 'Success!', data.message, 5000);
    }
    return false;
}

async function showDialog(data) {
    const dialogResponse = await secureFetch(`/dialog_${data.dialogId}`)
    var dialogId = data.dialogId + '-' + Date.now(); // Generate a unique ID
    const newDialog = $(dialogResponse.data)

    newDialog.filter('div').first().attr('id', dialogId);
    $(`#${dialogId}`).remove();
    $('body').append(newDialog);

    const dialog = $(`#${dialogId}`)
    dialog.modal('show');
    initDialog(dialog)
    $('.modal-backdrop').remove();

    const dialogForm = dialog.find(`#dialogForm`)
    dialog.find('#submitReason').on('click', async function(event) {
        event.preventDefault();
        if(!isValidDialogInput(dialog)) {
            return
        }

        var formDataArray = dialogForm.serializeArray();
        var formData = {};
        $.each(formDataArray, function(_, field) {
            formData[field.name] = field.value;
        });
    
        const dialogFormResponse = await secureFetch(data.dialogAction,  {
            method: 'POST',
            body: JSON.stringify(formData)
        });
        if(dialogFormResponse.ok) {
            dialog.modal('hide');
            dialog.remove();                
        }
    });
    dialog.on('hidden.bs.modal', function() {
        $(this).remove();
    });
}

function navigateTo(modelType) {
    window.location.href = `/model/${modelType}`;
}

function toggleSidebar() {
    document.getElementById("sidebar").classList.toggle('active');
    document.getElementById("content").classList.toggle('active');
}

function adjustBackButton() {
    const backBtn = $('#backBtn');
    if (window.history.length > 1) {
        backBtn.attr("disabled", false);
        backBtn.on('click', function () {
            window.history.back()
        });
    } else {
        backBtn.attr("disabled", true);
        backBtn.css("background-color", 'gray');
    }
}

function selectNavigationItem() {
    const currentPath = window.location.pathname.replace(/\/$/, '');
    $('#sidebar li a').each(function(idx, element) {
        const $element = $(element)
        if ($element.children().length > 0 || $.trim($element.html()) === "") {
            return;
        }
        console.log($element)
        const linkPath = $element.attr('href').replace(/\/$/, '');
        if (linkPath === currentPath) {
            $element.parent().addClass('active');
            const parentMenu = $element.closest('.collapse')[0];
            if(parentMenu !== undefined) new bootstrap.Collapse(parentMenu, {toggle: true});
        }
    });
}

$(document).ready(function() {
    adjustBackButton();
    selectNavigationItem();
})