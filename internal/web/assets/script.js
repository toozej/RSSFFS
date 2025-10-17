// DOM elements
const form = document.getElementById('rss-form');
const urlInput = document.getElementById('url');
const categorySelect = document.getElementById('category');
const submitBtn = document.getElementById('submit-btn');
const urlError = document.getElementById('url-error');
const categoryError = document.getElementById('category-error');
const categoryLoading = document.getElementById('category-loading');
const toastContainer = document.getElementById('toast-container');

// Form validation state
let isSubmitting = false;

// Helper to get a cookie value by name
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

// Initialize form validation
document.addEventListener('DOMContentLoaded', function() {
    setupFormValidation();
    setupFormSubmission();
    loadCategories();
});

// Setup real-time form validation
function setupFormValidation() {
    // URL validation
    urlInput.addEventListener('input', validateURL);
    urlInput.addEventListener('blur', validateURL);
    
    // Category validation (now for select)
    categorySelect.addEventListener('change', validateCategory);
}

// URL validation function
function validateURL() {
    const url = urlInput.value.trim();
    const urlPattern = /^https?:\/\/.+\..+/i;
    
    clearError(urlError);
    
    if (!url) {
        showError(urlError, 'URL is required');
        return false;
    }
    
    if (!urlPattern.test(url)) {
        showError(urlError, 'Please enter a valid URL (e.g., https://example.com)');
        return false;
    }
    
    if (url.length > 2048) {
        showError(urlError, 'URL is too long (maximum 2048 characters)');
        return false;
    }
    
    return true;
}

// Category validation function
function validateCategory() {
    clearError(categoryError);
    // Category is optional and pre-validated from server, so always valid
    return true;
}

// Load categories from server
async function loadCategories() {
    try {
        showCategoryLoading(true);
        
        const response = await fetch('/categories', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        
        if (data.success && data.categories) {
            populateCategories(data.categories);
            
            // Check if we're using fallback categories (ID = 0 indicates fallback)
            const usingFallback = data.categories.length > 0 && data.categories[0].id === 0;
            if (usingFallback) {
                console.info('Using fallback categories - RSS reader not accessible');
            }
        } else {
            throw new Error(data.error || 'Failed to load categories');
        }
        
    } catch (error) {
        console.error('Error loading categories:', error);
        showError(categoryError, 'Failed to load categories. You can still submit without selecting a category.');
    } finally {
        showCategoryLoading(false);
    }
}

// Populate category dropdown
function populateCategories(categories) {
    // Clear existing options except the first one
    while (categorySelect.children.length > 1) {
        categorySelect.removeChild(categorySelect.lastChild);
    }
    
    // Add categories
    categories.forEach(category => {
        const option = document.createElement('option');
        option.value = category.title;
        option.textContent = category.title;
        categorySelect.appendChild(option);
    });
    
    // Enable the select
    categorySelect.disabled = false;
}

// Show/hide category loading state
function showCategoryLoading(loading) {
    if (loading) {
        categoryLoading.style.display = 'block';
        categorySelect.disabled = true;
    } else {
        categoryLoading.style.display = 'none';
        categorySelect.disabled = false;
    }
}

// Show validation error
function showError(errorElement, message) {
    errorElement.textContent = message;
    errorElement.style.display = 'block';
}

// Clear validation error
function clearError(errorElement) {
    errorElement.textContent = '';
    errorElement.style.display = 'none';
}

// Setup form submission
function setupFormSubmission() {
    form.addEventListener('submit', handleFormSubmit);
}

// Handle form submission
async function handleFormSubmit(event) {
    event.preventDefault();
    
    // Prevent duplicate submissions
    if (isSubmitting) {
        return;
    }
    
    // Validate all fields
    const isURLValid = validateURL();
    const isCategoryValid = validateCategory();
    
    if (!isURLValid || !isCategoryValid) {
        showToast('Please fix the validation errors', 'error');
        return;
    }
    
    // Start submission process
    setSubmittingState(true);
    
    try {
        const singleUrlModeCheckbox = document.getElementById('single-url-mode');
        const formData = {
            url: urlInput.value.trim(),
            category: categorySelect.value.trim(),
            single_url_mode: singleUrlModeCheckbox ? singleUrlModeCheckbox.checked : false
        };
        
        const response = await submitForm(formData);
        handleSubmissionResponse(response);
        
    } catch (error) {
        console.error('Form submission error:', error);
        showToast('Network error. Please check your connection and try again.', 'error');
    } finally {
        setSubmittingState(false);
    }
}

// Submit form data to server
async function submitForm(formData) {
    const body = new URLSearchParams();
    body.append('url', formData.url);
    body.append('category', formData.category);
    body.append('single_url_mode', formData.single_url_mode ? 'true' : 'false');

    const csrfToken = getCookie('csrf_token');
    if (!csrfToken) {
        showToast('Security token missing. Please refresh the page.', 'error');
        throw new Error('CSRF token not found');
    }

    const response = await fetch('/submit', {
        method: 'POST',
        body: body,
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'X-CSRF-Token': csrfToken
        }
    });

    if (!response.ok) {
        if (response.status === 403) {
            showToast('Security token invalid. Please refresh the page and try again.', 'error');
        } else if (response.status === 400) {
            try {
                const errData = await response.json();
                const message = errData.message || 'Invalid data submitted. Please check the form.';
                showToast(message, 'error');
            } catch (e) {
                showToast('Invalid data submitted. Please check the form.', 'error');
            }
        }
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
}

// Handle submission response
function handleSubmissionResponse(response) {
    if (response.success) {
        const message = response.count > 0 
            ? `Success! Found and added ${response.count} RSS feed${response.count !== 1 ? 's' : ''}`
            : 'Success! RSS feed processing completed';
        
        showToast(message, 'success');
        
        // Clear form on success
        form.reset();
        clearError(urlError);
        clearError(categoryError);
        
    } else {
        const errorMessage = response.message || response.error || 'An error occurred while processing your request';
        showToast(errorMessage, 'error');
    }
}

// Set form submitting state
function setSubmittingState(submitting) {
    isSubmitting = submitting;
    submitBtn.disabled = submitting;
    
    if (submitting) {
        submitBtn.classList.add('loading');
    } else {
        submitBtn.classList.remove('loading');
    }
}

// Toast notification system
function showToast(message, type = 'info', duration = 5000) {
    const toast = createToastElement(message, type);
    toastContainer.appendChild(toast);
    
    // Trigger animation
    setTimeout(() => {
        toast.classList.add('show');
    }, 10);
    
    // Auto-dismiss
    setTimeout(() => {
        dismissToast(toast);
    }, duration);
    
    // Click to dismiss
    toast.addEventListener('click', () => {
        dismissToast(toast);
    });
}

// Create toast element
function createToastElement(message, type) {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    toast.style.cursor = 'pointer';
    toast.title = 'Click to dismiss';
    return toast;
}

// Dismiss toast notification
function dismissToast(toast) {
    toast.classList.remove('show');
    
    setTimeout(() => {
        if (toast.parentNode) {
            toast.parentNode.removeChild(toast);
        }
    }, 300);
}

// Keyboard accessibility
document.addEventListener('keydown', function(event) {
    // Escape key dismisses all toasts
    if (event.key === 'Escape') {
        const toasts = document.querySelectorAll('.toast');
        toasts.forEach(dismissToast);
    }
});

// Handle network connectivity
window.addEventListener('online', function() {
    showToast('Connection restored', 'success', 3000);
});

window.addEventListener('offline', function() {
    showToast('Connection lost. Please check your internet connection.', 'error', 8000);
});

// Form auto-save to localStorage (optional enhancement)
function saveFormData() {
    const singleUrlModeCheckbox = document.getElementById('single-url-mode');
    const formData = {
        url: urlInput.value,
        category: categorySelect.value,
        single_url_mode: singleUrlModeCheckbox ? singleUrlModeCheckbox.checked : false
    };
    localStorage.setItem('rss-form-data', JSON.stringify(formData));
}

function loadFormData() {
    try {
        const savedData = localStorage.getItem('rss-form-data');
        if (savedData) {
            const formData = JSON.parse(savedData);
            if (formData.url) urlInput.value = formData.url;
            if (formData.category) categorySelect.value = formData.category;
            
            const singleUrlModeCheckbox = document.getElementById('single-url-mode');
            if (singleUrlModeCheckbox && typeof formData.single_url_mode === 'boolean') {
                singleUrlModeCheckbox.checked = formData.single_url_mode;
            }
        }
    } catch (error) {
        console.warn('Could not load saved form data:', error);
    }
}

// Save form data on input
urlInput.addEventListener('input', saveFormData);
categorySelect.addEventListener('change', saveFormData);

// Add event listener for checkbox when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    const singleUrlModeCheckbox = document.getElementById('single-url-mode');
    if (singleUrlModeCheckbox) {
        singleUrlModeCheckbox.addEventListener('change', saveFormData);
    }
});

// Load saved form data on page load
document.addEventListener('DOMContentLoaded', loadFormData);

// Clear saved data on successful submission
function clearSavedFormData() {
    localStorage.removeItem('rss-form-data');
}

// Update handleSubmissionResponse to clear saved data on success
const originalHandleSubmissionResponse = handleSubmissionResponse;
handleSubmissionResponse = function(response) {
    originalHandleSubmissionResponse(response);
    if (response.success) {
        clearSavedFormData();
    }
};