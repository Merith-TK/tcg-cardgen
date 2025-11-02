// Canvas rendering system for the cardstyle designer
// Handles layer rendering, mouse interactions, and visual feedback

class CanvasRenderer {
    constructor() {
        this.canvas = document.getElementById('designCanvas');
        this.overlayCanvas = document.getElementById('overlayCanvas');
        this.ctx = this.canvas.getContext('2d');
        this.overlayCtx = this.overlayCanvas.getContext('2d');
        
        this.zoom = 1.0;
        this.offsetX = 0;
        this.offsetY = 0;
        this.isDragging = false;
        this.isResizing = false;
        this.dragStartX = 0;
        this.dragStartY = 0;
        this.selectedLayers = new Set();
        this.hoveredLayer = null;
        this.resizeHandle = null;
        
        this.template = null;
        this.layers = [];
        
        this.setupEventListeners();
        this.render();
    }
    
    setupEventListeners() {
        // Mouse events on overlay canvas
        this.overlayCanvas.addEventListener('mousedown', this.onMouseDown.bind(this));
        this.overlayCanvas.addEventListener('mousemove', this.onMouseMove.bind(this));
        this.overlayCanvas.addEventListener('mouseup', this.onMouseUp.bind(this));
        this.overlayCanvas.addEventListener('wheel', this.onWheel.bind(this));
        
        // Prevent context menu
        this.overlayCanvas.addEventListener('contextmenu', (e) => e.preventDefault());
        
        // Canvas resize observer
        const resizeObserver = new ResizeObserver(() => {
            this.updateCanvasSize();
        });
        resizeObserver.observe(this.canvas.parentElement);
    }
    
    updateCanvasSize() {
        const container = this.canvas.parentElement;
        const rect = container.getBoundingClientRect();
        
        // Set canvas size to fit container while maintaining aspect ratio
        const cardAspect = 750 / 1050; // TCG card aspect ratio
        const containerAspect = rect.width / rect.height;
        
        let canvasWidth, canvasHeight;
        if (containerAspect > cardAspect) {
            // Container is wider, fit to height
            canvasHeight = Math.min(rect.height - 40, 600); // Leave some margin
            canvasWidth = canvasHeight * cardAspect;
        } else {
            // Container is taller, fit to width
            canvasWidth = Math.min(rect.width - 40, 450); // Leave some margin
            canvasHeight = canvasWidth / cardAspect;
        }
        
        // Update canvas display size
        this.canvas.style.width = canvasWidth + 'px';
        this.canvas.style.height = canvasHeight + 'px';
        this.overlayCanvas.style.width = canvasWidth + 'px';
        this.overlayCanvas.style.height = canvasHeight + 'px';
        
        // Calculate zoom to fit card in canvas
        this.zoom = Math.min(canvasWidth / 750, canvasHeight / 1050);
        
        this.render();
    }
    
    setTemplate(template) {
        this.template = template;
        this.layers = template ? [...template.layers] : [];
        this.selectedLayers.clear();
        this.render();
    }
    
    addLayer(layer) {
        layer.id = layer.id || Utils.generateId();
        this.layers.push(layer);
        this.render();
        return layer;
    }
    
    updateLayer(layerId, updates) {
        const layer = this.layers.find(l => l.id === layerId);
        if (layer) {
            Object.assign(layer, updates);
            this.render();
        }
    }
    
    removeLayer(layerId) {
        const index = this.layers.findIndex(l => l.id === layerId);
        if (index >= 0) {
            this.layers.splice(index, 1);
            this.selectedLayers.delete(layerId);
            this.render();
        }
    }
    
    selectLayer(layerId, multi = false) {
        if (!multi) {
            this.selectedLayers.clear();
        }
        
        if (layerId) {
            this.selectedLayers.add(layerId);
        }
        
        this.render();
        
        // Notify layer panel
        if (window.LayerManager) {
            window.LayerManager.updateSelection(Array.from(this.selectedLayers));
        }
        
        // Notify properties panel
        if (window.PropertiesManager) {
            const selectedLayer = layerId ? this.layers.find(l => l.id === layerId) : null;
            window.PropertiesManager.setSelectedLayer(selectedLayer);
        }
    }
    
    getLayerAtPoint(x, y) {
        // Convert screen coordinates to canvas coordinates
        const rect = this.overlayCanvas.getBoundingClientRect();
        const canvasX = (x - rect.left) / this.zoom;
        const canvasY = (y - rect.top) / this.zoom;
        
        // Check layers in reverse order (topmost first)
        for (let i = this.layers.length - 1; i >= 0; i--) {
            const layer = this.layers[i];
            if (Utils.pointInRect(canvasX, canvasY, layer.region)) {
                return layer;
            }
        }
        return null;
    }
    
    getResizeHandle(layerId, x, y) {
        const layer = this.layers.find(l => l.id === layerId);
        if (!layer) return null;
        
        const rect = this.overlayCanvas.getBoundingClientRect();
        const canvasX = (x - rect.left) / this.zoom;
        const canvasY = (y - rect.top) / this.zoom;
        
        const region = layer.region;
        const handleSize = 8 / this.zoom;
        
        // Check each corner and edge
        const handles = [
            { name: 'nw', x: region.x, y: region.y },
            { name: 'n', x: region.x + region.width / 2, y: region.y },
            { name: 'ne', x: region.x + region.width, y: region.y },
            { name: 'e', x: region.x + region.width, y: region.y + region.height / 2 },
            { name: 'se', x: region.x + region.width, y: region.y + region.height },
            { name: 's', x: region.x + region.width / 2, y: region.y + region.height },
            { name: 'sw', x: region.x, y: region.y + region.height },
            { name: 'w', x: region.x, y: region.y + region.height / 2 }
        ];
        
        for (const handle of handles) {
            if (Utils.distance(canvasX, canvasY, handle.x, handle.y) <= handleSize) {
                return handle.name;
            }
        }
        
        return null;
    }
    
    onMouseDown(e) {
        const rect = this.overlayCanvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        this.dragStartX = x;
        this.dragStartY = y;
        
        // Check if clicking on a resize handle
        if (this.selectedLayers.size === 1) {
            const selectedId = Array.from(this.selectedLayers)[0];
            const handle = this.getResizeHandle(selectedId, e.clientX, e.clientY);
            if (handle) {
                this.isResizing = true;
                this.resizeHandle = handle;
                this.overlayCanvas.style.cursor = this.getResizeCursor(handle);
                return;
            }
        }
        
        // Check if clicking on a layer
        const layer = this.getLayerAtPoint(e.clientX, e.clientY);
        if (layer) {
            this.selectLayer(layer.id, e.ctrlKey || e.metaKey);
            this.isDragging = true;
            this.overlayCanvas.style.cursor = 'move';
        } else {
            // Clicked on empty space
            this.selectLayer(null);
        }
    }
    
    onMouseMove(e) {
        const rect = this.overlayCanvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        // Update cursor position display
        const canvasX = Math.round(x / this.zoom);
        const canvasY = Math.round(y / this.zoom);
        const cursorPos = document.getElementById('cursorPos');
        if (cursorPos) {
            cursorPos.textContent = `x: ${canvasX}, y: ${canvasY}`;
        }
        
        if (this.isResizing && this.selectedLayers.size === 1) {
            const selectedId = Array.from(this.selectedLayers)[0];
            const layer = this.layers.find(l => l.id === selectedId);
            if (layer) {
                this.handleResize(layer, x, y);
            }
        } else if (this.isDragging && this.selectedLayers.size > 0) {
            const deltaX = (x - this.dragStartX) / this.zoom;
            const deltaY = (y - this.dragStartY) / this.zoom;
            
            this.selectedLayers.forEach(layerId => {
                const layer = this.layers.find(l => l.id === layerId);
                if (layer) {
                    layer.region.x = Math.round(layer.region.x + deltaX);
                    layer.region.y = Math.round(layer.region.y + deltaY);
                }
            });
            
            this.dragStartX = x;
            this.dragStartY = y;
            this.render();
        } else {
            // Update hover state and cursor
            const layer = this.getLayerAtPoint(e.clientX, e.clientY);
            this.hoveredLayer = layer;
            
            if (layer && this.selectedLayers.has(layer.id)) {
                const handle = this.getResizeHandle(layer.id, e.clientX, e.clientY);
                if (handle) {
                    this.overlayCanvas.style.cursor = this.getResizeCursor(handle);
                } else {
                    this.overlayCanvas.style.cursor = 'move';
                }
            } else if (layer) {
                this.overlayCanvas.style.cursor = 'pointer';
            } else {
                this.overlayCanvas.style.cursor = 'default';
            }
            
            this.renderOverlay();
        }
    }
    
    onMouseUp(e) {
        this.isDragging = false;
        this.isResizing = false;
        this.resizeHandle = null;
        this.overlayCanvas.style.cursor = 'default';
        
        // Update properties panel with new values
        if (this.selectedLayers.size === 1 && window.PropertiesManager) {
            const selectedId = Array.from(this.selectedLayers)[0];
            const layer = this.layers.find(l => l.id === selectedId);
            window.PropertiesManager.setSelectedLayer(layer);
        }
    }
    
    onWheel(e) {
        e.preventDefault();
        
        const rect = this.overlayCanvas.getBoundingClientRect();
        const mouseX = e.clientX - rect.left;
        const mouseY = e.clientY - rect.top;
        
        const zoomFactor = e.deltaY > 0 ? 0.9 : 1.1;
        const newZoom = Utils.clamp(this.zoom * zoomFactor, 0.1, 5.0);
        
        // Update zoom level display
        const zoomLevel = document.getElementById('zoomLevel');
        if (zoomLevel) {
            zoomLevel.textContent = Math.round(newZoom * 100) + '%';
        }
        
        this.zoom = newZoom;
        this.render();
    }
    
    handleResize(layer, x, y) {
        const deltaX = (x - this.dragStartX) / this.zoom;
        const deltaY = (y - this.dragStartY) / this.zoom;
        
        const region = layer.region;
        const originalRegion = { ...region };
        
        switch (this.resizeHandle) {
            case 'nw':
                region.x += deltaX;
                region.y += deltaY;
                region.width -= deltaX;
                region.height -= deltaY;
                break;
            case 'n':
                region.y += deltaY;
                region.height -= deltaY;
                break;
            case 'ne':
                region.y += deltaY;
                region.width += deltaX;
                region.height -= deltaY;
                break;
            case 'e':
                region.width += deltaX;
                break;
            case 'se':
                region.width += deltaX;
                region.height += deltaY;
                break;
            case 's':
                region.height += deltaY;
                break;
            case 'sw':
                region.x += deltaX;
                region.width -= deltaX;
                region.height += deltaY;
                break;
            case 'w':
                region.x += deltaX;
                region.width -= deltaX;
                break;
        }
        
        // Enforce minimum size
        const minSize = 10;
        if (region.width < minSize) {
            region.width = minSize;
            region.x = originalRegion.x;
        }
        if (region.height < minSize) {
            region.height = minSize;
            region.y = originalRegion.y;
        }
        
        // Round to integers
        region.x = Math.round(region.x);
        region.y = Math.round(region.y);
        region.width = Math.round(region.width);
        region.height = Math.round(region.height);
        
        this.dragStartX = x;
        this.dragStartY = y;
        this.render();
    }
    
    getResizeCursor(handle) {
        const cursors = {
            'nw': 'nw-resize',
            'n': 'n-resize',
            'ne': 'ne-resize',
            'e': 'e-resize',
            'se': 'se-resize',
            's': 's-resize',
            'sw': 'sw-resize',
            'w': 'w-resize'
        };
        return cursors[handle] || 'default';
    }
    
    render() {
        this.renderCanvas();
        this.renderOverlay();
    }
    
    renderCanvas() {
        const ctx = this.ctx;
        ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        
        // Set white background
        ctx.fillStyle = '#ffffff';
        ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        
        // Apply zoom and offset
        ctx.save();
        ctx.scale(this.zoom, this.zoom);
        
        // Render each layer
        this.layers.forEach(layer => {
            this.renderLayer(ctx, layer);
        });
        
        ctx.restore();
    }
    
    renderLayer(ctx, layer) {
        if (!layer.region) return;
        
        const { x, y, width, height } = layer.region;
        
        ctx.save();
        
        switch (layer.type) {
            case 'shape':
                this.renderShapeLayer(ctx, layer);
                break;
            case 'text':
                this.renderTextLayer(ctx, layer);
                break;
            case 'image':
                this.renderImageLayer(ctx, layer);
                break;
        }
        
        ctx.restore();
    }
    
    renderShapeLayer(ctx, layer) {
        const { x, y, width, height } = layer.region;
        const fill = layer.fill || '#CCCCCCFF';
        const stroke = layer.stroke || '#000000FF';
        const strokeWidth = layer.stroke_width || 1;
        
        // Convert RGBA colors
        if (fill && fill !== 'none') {
            ctx.fillStyle = Utils.hexToRgba(fill);
        }
        if (stroke && stroke !== 'none') {
            ctx.strokeStyle = Utils.hexToRgba(stroke);
            ctx.lineWidth = strokeWidth;
        }
        
        ctx.beginPath();
        
        switch (layer.shape) {
            case 'rectangle':
                if (layer.corner_radius && layer.corner_radius > 0) {
                    this.drawRoundedRect(ctx, x, y, width, height, layer.corner_radius);
                } else {
                    ctx.rect(x, y, width, height);
                }
                break;
            case 'circle':
                const radius = Math.min(width, height) / 2;
                ctx.arc(x + width / 2, y + height / 2, radius, 0, 2 * Math.PI);
                break;
            case 'ellipse':
                ctx.ellipse(x + width / 2, y + height / 2, width / 2, height / 2, 0, 0, 2 * Math.PI);
                break;
            case 'polygon':
                if (layer.points && layer.points.length >= 3) {
                    layer.points.forEach((point, index) => {
                        const px = x + (point[0] / 100) * width;
                        const py = y + (point[1] / 100) * height;
                        if (index === 0) {
                            ctx.moveTo(px, py);
                        } else {
                            ctx.lineTo(px, py);
                        }
                    });
                    ctx.closePath();
                }
                break;
        }
        
        if (fill && fill !== 'none') {
            ctx.fill();
        }
        if (stroke && stroke !== 'none') {
            ctx.stroke();
        }
    }
    
    renderTextLayer(ctx, layer) {
        const { x, y, width, height } = layer.region;
        const content = layer.content || 'Text';
        const font = layer.font || {};
        
        // Set font
        const fontSize = font.size || 14;
        const fontFamily = font.family || 'Arial';
        const fontWeight = font.weight || 'normal';
        const fontStyle = font.style || 'normal';
        
        ctx.font = `${fontStyle} ${fontWeight} ${fontSize}px ${fontFamily}`;
        ctx.fillStyle = Utils.hexToRgba(font.color || '#000000FF');
        ctx.textBaseline = 'top';
        
        // Simple text alignment
        const align = layer.align || 'left';
        switch (align) {
            case 'center':
                ctx.textAlign = 'center';
                ctx.fillText(content, x + width / 2, y);
                break;
            case 'right':
                ctx.textAlign = 'right';
                ctx.fillText(content, x + width, y);
                break;
            default:
                ctx.textAlign = 'left';
                ctx.fillText(content, x, y);
                break;
        }
    }
    
    renderImageLayer(ctx, layer) {
        const { x, y, width, height } = layer.region;
        
        // For now, render a placeholder
        ctx.fillStyle = '#f0f0f0';
        ctx.fillRect(x, y, width, height);
        ctx.strokeStyle = '#ccc';
        ctx.lineWidth = 1;
        ctx.strokeRect(x, y, width, height);
        
        // Draw image icon
        ctx.fillStyle = '#999';
        ctx.font = '24px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText('🖼️', x + width / 2, y + height / 2);
    }
    
    drawRoundedRect(ctx, x, y, width, height, radius) {
        ctx.moveTo(x + radius, y);
        ctx.lineTo(x + width - radius, y);
        ctx.quadraticCurveTo(x + width, y, x + width, y + radius);
        ctx.lineTo(x + width, y + height - radius);
        ctx.quadraticCurveTo(x + width, y + height, x + width - radius, y + height);
        ctx.lineTo(x + radius, y + height);
        ctx.quadraticCurveTo(x, y + height, x, y + height - radius);
        ctx.lineTo(x, y + radius);
        ctx.quadraticCurveTo(x, y, x + radius, y);
    }
    
    renderOverlay() {
        const ctx = this.overlayCtx;
        ctx.clearRect(0, 0, this.overlayCanvas.width, this.overlayCanvas.height);
        
        ctx.save();
        ctx.scale(this.zoom, this.zoom);
        
        // Render selection outlines
        this.selectedLayers.forEach(layerId => {
            const layer = this.layers.find(l => l.id === layerId);
            if (layer) {
                this.renderSelectionOutline(ctx, layer);
            }
        });
        
        // Render hover outline
        if (this.hoveredLayer && !this.selectedLayers.has(this.hoveredLayer.id)) {
            this.renderHoverOutline(ctx, this.hoveredLayer);
        }
        
        ctx.restore();
    }
    
    renderSelectionOutline(ctx, layer) {
        const { x, y, width, height } = layer.region;
        
        // Selection outline
        ctx.strokeStyle = '#3498db';
        ctx.lineWidth = 2 / this.zoom;
        ctx.setLineDash([5 / this.zoom, 5 / this.zoom]);
        ctx.strokeRect(x - 1, y - 1, width + 2, height + 2);
        ctx.setLineDash([]);
        
        // Resize handles
        const handleSize = 8 / this.zoom;
        const handles = [
            { x: x, y: y },
            { x: x + width / 2, y: y },
            { x: x + width, y: y },
            { x: x + width, y: y + height / 2 },
            { x: x + width, y: y + height },
            { x: x + width / 2, y: y + height },
            { x: x, y: y + height },
            { x: x, y: y + height / 2 }
        ];
        
        ctx.fillStyle = '#3498db';
        ctx.strokeStyle = '#ffffff';
        ctx.lineWidth = 1 / this.zoom;
        
        handles.forEach(handle => {
            ctx.fillRect(handle.x - handleSize / 2, handle.y - handleSize / 2, handleSize, handleSize);
            ctx.strokeRect(handle.x - handleSize / 2, handle.y - handleSize / 2, handleSize, handleSize);
        });
    }
    
    renderHoverOutline(ctx, layer) {
        const { x, y, width, height } = layer.region;
        
        ctx.strokeStyle = '#95a5a6';
        ctx.lineWidth = 1 / this.zoom;
        ctx.setLineDash([3 / this.zoom, 3 / this.zoom]);
        ctx.strokeRect(x, y, width, height);
        ctx.setLineDash([]);
    }
}

// Make CanvasRenderer available globally
window.CanvasRenderer = CanvasRenderer;