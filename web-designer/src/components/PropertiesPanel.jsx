import React from 'react'

function PropertiesPanel({ selectedLayer, onUpdateLayer }) {
  if (!selectedLayer) {
    return (
      <div className="panel properties-panel" style={{ width: '300px' }}>
        <div className="panel-header">
          <h3>🎛️ Properties</h3>
        </div>
        <div className="panel-content">
          <p className="empty-state">Select a layer to edit its properties</p>
        </div>
      </div>
    )
  }

  const updateProperty = (property, value) => {
    if (property.includes('.')) {
      const [parent, child] = property.split('.')
      onUpdateLayer(selectedLayer.id, {
        [parent]: {
          ...selectedLayer[parent],
          [child]: value
        }
      })
    } else {
      onUpdateLayer(selectedLayer.id, { [property]: value })
    }
  }

  const renderTextProperties = () => (
    <div className="property-section">
      <h4>Text Properties</h4>
      
      <div className="form-group">
        <label className="form-label">Content</label>
        <textarea
          className="form-control"
          value={selectedLayer.content || ''}
          onChange={(e) => updateProperty('content', e.target.value)}
          rows="2"
        />
      </div>
      
      <div className="form-group">
        <label className="form-label">Font Family</label>
        <select
          className="form-control"
          value={selectedLayer.font?.family || 'Arial'}
          onChange={(e) => updateProperty('font.family', e.target.value)}
        >
          <option value="Arial">Arial</option>
          <option value="Helvetica">Helvetica</option>
          <option value="Times New Roman">Times New Roman</option>
          <option value="Georgia">Georgia</option>
          <option value="Verdana">Verdana</option>
        </select>
      </div>
      
      <div className="form-group">
        <label className="form-label">Font Size</label>
        <input
          type="number"
          className="form-control"
          value={selectedLayer.font?.size || 16}
          onChange={(e) => updateProperty('font.size', parseInt(e.target.value))}
          min="8"
          max="72"
        />
      </div>
      
      <div className="form-group">
        <label className="form-label">Font Weight</label>
        <select
          className="form-control"
          value={selectedLayer.font?.weight || 'normal'}
          onChange={(e) => updateProperty('font.weight', e.target.value)}
        >
          <option value="normal">Normal</option>
          <option value="bold">Bold</option>
          <option value="lighter">Lighter</option>
        </select>
      </div>
      
      <div className="form-group">
        <label className="form-label">Color</label>
        <input
          type="color"
          className="form-control"
          value={selectedLayer.font?.color?.substring(0, 7) || '#000000'}
          onChange={(e) => updateProperty('font.color', e.target.value + 'FF')}
        />
      </div>
      
      <div className="form-group">
        <label className="form-label">Alignment</label>
        <select
          className="form-control"
          value={selectedLayer.align || 'left'}
          onChange={(e) => updateProperty('align', e.target.value)}
        >
          <option value="left">Left</option>
          <option value="center">Center</option>
          <option value="right">Right</option>
        </select>
      </div>
    </div>
  )

  const renderShapeProperties = () => (
    <div className="property-section">
      <h4>Shape Properties</h4>
      
      <div className="form-group">
        <label className="form-label">Shape Type</label>
        <select
          className="form-control"
          value={selectedLayer.shape || 'rectangle'}
          onChange={(e) => updateProperty('shape', e.target.value)}
        >
          <option value="rectangle">Rectangle</option>
          <option value="circle">Circle</option>
          <option value="ellipse">Ellipse</option>
        </select>
      </div>
      
      <div className="form-group">
        <label className="form-label">Fill Color</label>
        <input
          type="color"
          className="form-control"
          value={selectedLayer.fill?.substring(0, 7) || '#CCCCCC'}
          onChange={(e) => updateProperty('fill', e.target.value + 'FF')}
        />
      </div>
      
      <div className="form-group">
        <label className="form-label">Stroke Color</label>
        <input
          type="color"
          className="form-control"
          value={selectedLayer.stroke?.substring(0, 7) || '#000000'}
          onChange={(e) => updateProperty('stroke', e.target.value + 'FF')}
        />
      </div>
      
      <div className="form-group">
        <label className="form-label">Stroke Width</label>
        <input
          type="number"
          className="form-control"
          value={selectedLayer.stroke_width || 1}
          onChange={(e) => updateProperty('stroke_width', parseInt(e.target.value))}
          min="0"
          max="10"
        />
      </div>
      
      {selectedLayer.shape === 'rectangle' && (
        <div className="form-group">
          <label className="form-label">Corner Radius</label>
          <input
            type="number"
            className="form-control"
            value={selectedLayer.corner_radius || 0}
            onChange={(e) => updateProperty('corner_radius', parseInt(e.target.value))}
            min="0"
            max="50"
          />
        </div>
      )}
    </div>
  )

  const renderImageProperties = () => (
    <div className="property-section">
      <h4>Image Properties</h4>
      
      <div className="form-group">
        <label className="form-label">Image URL</label>
        <input
          type="text"
          className="form-control"
          value={selectedLayer.src || ''}
          onChange={(e) => updateProperty('src', e.target.value)}
          placeholder="Enter image URL or path"
        />
      </div>
      
      <div className="form-group">
        <label className="form-label">Upload Image</label>
        <input
          type="file"
          className="form-control"
          accept="image/*"
          onChange={(e) => {
            const file = e.target.files[0]
            if (file) {
              const reader = new FileReader()
              reader.onload = (event) => {
                updateProperty('src', event.target.result)
              }
              reader.readAsDataURL(file)
            }
          }}
        />
      </div>
    </div>
  )

  return (
    <div className="panel properties-panel" style={{ width: '300px' }}>
      <div className="panel-header">
        <h3>🎛️ Properties</h3>
        <span className="layer-type-badge">{selectedLayer.type}</span>
      </div>
      
      <div className="panel-content">
        {/* Basic Properties */}
        <div className="property-section">
          <h4>Basic Properties</h4>
          
          <div className="form-group">
            <label className="form-label">Name</label>
            <input
              type="text"
              className="form-control"
              value={selectedLayer.name}
              onChange={(e) => updateProperty('name', e.target.value)}
            />
          </div>
          
          <div className="form-group">
            <label className="form-label">X Position</label>
            <input
              type="number"
              className="form-control"
              value={selectedLayer.region?.x || 0}
              onChange={(e) => updateProperty('region.x', parseFloat(e.target.value))}
            />
          </div>
          
          <div className="form-group">
            <label className="form-label">Y Position</label>
            <input
              type="number"
              className="form-control"
              value={selectedLayer.region?.y || 0}
              onChange={(e) => updateProperty('region.y', parseFloat(e.target.value))}
            />
          </div>
          
          <div className="form-group">
            <label className="form-label">Width</label>
            <input
              type="number"
              className="form-control"
              value={selectedLayer.region?.width || 100}
              onChange={(e) => updateProperty('region.width', parseFloat(e.target.value))}
              min="1"
            />
          </div>
          
          <div className="form-group">
            <label className="form-label">Height</label>
            <input
              type="number"
              className="form-control"
              value={selectedLayer.region?.height || 100}
              onChange={(e) => updateProperty('region.height', parseFloat(e.target.value))}
              min="1"
            />
          </div>
        </div>

        {/* Type-specific Properties */}
        {selectedLayer.type === 'text' && renderTextProperties()}
        {selectedLayer.type === 'shape' && renderShapeProperties()}
        {selectedLayer.type === 'image' && renderImageProperties()}
      </div>
    </div>
  )
}

export default PropertiesPanel