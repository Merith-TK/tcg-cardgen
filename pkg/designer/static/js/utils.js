// Utility functions for the cardstyle designer
// Zero external dependencies - all vanilla JavaScript

/**
 * Generate a unique ID for layers
 */
function generateId() {
    return 'layer_' + Math.random().toString(36).substr(2, 9);
}

/**
 * Parse RGBA color string to components
 * @param {string} colorStr - Color string like "#FF0000FF"
 * @returns {object} - {r, g, b, a} components
 */
function parseRGBA(colorStr) {
    if (!colorStr || !colorStr.startsWith('#')) {
        return { r: 0, g: 0, b: 0, a: 255 };
    }
    
    const hex = colorStr.slice(1);
    let r, g, b, a = 255;
    
    switch (hex.length) {
        case 3: // #RGB
            r = parseInt(hex[0], 16) * 17;
            g = parseInt(hex[1], 16) * 17;
            b = parseInt(hex[2], 16) * 17;
            break;
        case 4: // #RGBA
            r = parseInt(hex[0], 16) * 17;
            g = parseInt(hex[1], 16) * 17;
            b = parseInt(hex[2], 16) * 17;
            a = parseInt(hex[3], 16) * 17;
            break;
        case 6: // #RRGGBB
            r = parseInt(hex.substr(0, 2), 16);
            g = parseInt(hex.substr(2, 2), 16);
            b = parseInt(hex.substr(4, 2), 16);
            break;
        case 8: // #RRGGBBAA
            r = parseInt(hex.substr(0, 2), 16);
            g = parseInt(hex.substr(2, 2), 16);
            b = parseInt(hex.substr(4, 2), 16);
            a = parseInt(hex.substr(6, 2), 16);
            break;
        default:
            return { r: 0, g: 0, b: 0, a: 255 };
    }
    
    return { r, g, b, a };
}

/**
 * Convert RGBA components to color string
 * @param {number} r - Red (0-255)
 * @param {number} g - Green (0-255)
 * @param {number} b - Blue (0-255)
 * @param {number} a - Alpha (0-255)
 * @returns {string} - Color string like "#FF0000FF"
 */
function rgbaToHex(r, g, b, a = 255) {
    const toHex = (n) => Math.round(Math.max(0, Math.min(255, n))).toString(16).padStart(2, '0');
    return `#${toHex(r)}${toHex(g)}${toHex(b)}${toHex(a)}`;
}

/**
 * Convert RGBA hex to CSS rgba() string
 * @param {string} hexColor - Color like "#FF0000FF"
 * @returns {string} - CSS rgba string like "rgba(255, 0, 0, 1)"
 */
function hexToRgba(hexColor) {
    const { r, g, b, a } = parseRGBA(hexColor);
    return `rgba(${r}, ${g}, ${b}, ${a / 255})`;
}

/**
 * Clamp a value between min and max
 * @param {number} value
 * @param {number} min
 * @param {number} max
 * @returns {number}
 */
function clamp(value, min, max) {
    return Math.max(min, Math.min(max, value));
}

/**
 * Check if point is inside a rectangle
 * @param {number} x
 * @param {number} y
 * @param {object} rect - {x, y, width, height}
 * @returns {boolean}
 */
function pointInRect(x, y, rect) {
    return x >= rect.x && x <= rect.x + rect.width &&
           y >= rect.y && y <= rect.y + rect.height;
}

/**
 * Get the distance between two points
 * @param {number} x1
 * @param {number} y1
 * @param {number} x2
 * @param {number} y2
 * @returns {number}
 */
function distance(x1, y1, x2, y2) {
    return Math.sqrt((x2 - x1) ** 2 + (y2 - y1) ** 2);
}

/**
 * Debounce function calls
 * @param {function} func
 * @param {number} wait
 * @returns {function}
 */
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

/**
 * Show status message
 * @param {string} message
 * @param {string} type - 'info', 'success', 'warning', 'error'
 */
function showStatus(message, type = 'info') {
    const statusText = document.getElementById('statusText');
    if (statusText) {
        statusText.textContent = message;
        statusText.className = `status-${type}`;
        
        // Clear after 3 seconds
        setTimeout(() => {
            statusText.textContent = 'Ready';
            statusText.className = '';
        }, 3000);
    }
}

/**
 * Create a deep copy of an object
 * @param {object} obj
 * @returns {object}
 */
function deepClone(obj) {
    if (obj === null || typeof obj !== 'object') return obj;
    if (obj instanceof Date) return new Date(obj.getTime());
    if (obj instanceof Array) return obj.map(item => deepClone(item));
    if (typeof obj === 'object') {
        const copy = {};
        Object.keys(obj).forEach(key => {
            copy[key] = deepClone(obj[key]);
        });
        return copy;
    }
}

/**
 * Format file size in human readable format
 * @param {number} bytes
 * @returns {string}
 */
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * Download a file with given content
 * @param {string} content
 * @param {string} filename
 * @param {string} mimeType
 */
function downloadFile(content, filename, mimeType = 'text/plain') {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}

/**
 * Make an API request
 * @param {string} url
 * @param {object} options
 * @returns {Promise}
 */
async function apiRequest(url, options = {}) {
    try {
        const response = await fetch(url, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('API Request failed:', error);
        showStatus(`API Error: ${error.message}`, 'error');
        throw error;
    }
}

/**
 * Validate template data
 * @param {object} template
 * @returns {object} - {valid: boolean, errors: string[]}
 */
function validateTemplate(template) {
    const errors = [];
    
    if (!template.name || template.name.trim() === '') {
        errors.push('Template name is required');
    }
    
    if (!template.tcg || template.tcg.trim() === '') {
        errors.push('TCG type is required');
    }
    
    if (!template.dimensions || !template.dimensions.width || !template.dimensions.height) {
        errors.push('Template dimensions are required');
    }
    
    if (!template.layers || !Array.isArray(template.layers)) {
        errors.push('Template must have layers array');
    } else if (template.layers.length === 0) {
        errors.push('Template must have at least one layer');
    }
    
    // Validate layers
    template.layers.forEach((layer, index) => {
        if (!layer.name || layer.name.trim() === '') {
            errors.push(`Layer ${index + 1} must have a name`);
        }
        
        if (!layer.type || !['text', 'shape', 'image'].includes(layer.type)) {
            errors.push(`Layer ${index + 1} must have a valid type (text, shape, or image)`);
        }
        
        if (!layer.region || typeof layer.region.x !== 'number' || 
            typeof layer.region.y !== 'number' || typeof layer.region.width !== 'number' || 
            typeof layer.region.height !== 'number') {
            errors.push(`Layer ${index + 1} must have valid region (x, y, width, height)`);
        }
    });
    
    return {
        valid: errors.length === 0,
        errors
    };
}

/**
 * Convert template to YAML string
 * @param {object} template
 * @returns {string}
 */
function templateToYAML(template) {
    // Simple YAML serialization (for a full implementation, you'd want a proper YAML library)
    // This is a basic implementation for demonstration
    
    function yamlValue(value, indent = 0) {
        const prefix = '  '.repeat(indent);
        
        if (value === null || value === undefined) {
            return 'null';
        }
        
        if (typeof value === 'string') {
            // Simple string escaping
            if (value.includes('\n') || value.includes('"') || value.includes("'")) {
                return `"${value.replace(/"/g, '\\"')}"`;
            }
            return value;
        }
        
        if (typeof value === 'number' || typeof value === 'boolean') {
            return value.toString();
        }
        
        if (Array.isArray(value)) {
            if (value.length === 0) return '[]';
            return '\n' + value.map(item => prefix + '- ' + yamlValue(item, indent + 1).replace(/^\s+/, '')).join('\n');
        }
        
        if (typeof value === 'object') {
            if (Object.keys(value).length === 0) return '{}';
            return '\n' + Object.entries(value).map(([key, val]) => {
                const yamlVal = yamlValue(val, indent + 1);
                if (yamlVal.startsWith('\n')) {
                    return prefix + key + ':' + yamlVal;
                } else {
                    return prefix + key + ': ' + yamlVal;
                }
            }).join('\n');
        }
        
        return value.toString();
    }
    
    return Object.entries(template).map(([key, value]) => {
        const yamlVal = yamlValue(value);
        if (yamlVal.startsWith('\n')) {
            return key + ':' + yamlVal;
        } else {
            return key + ': ' + yamlVal;
        }
    }).join('\n');
}

// Export utilities for global use
window.Utils = {
    generateId,
    parseRGBA,
    rgbaToHex,
    hexToRgba,
    clamp,
    pointInRect,
    distance,
    debounce,
    showStatus,
    deepClone,
    formatFileSize,
    downloadFile,
    apiRequest,
    validateTemplate,
    templateToYAML
};