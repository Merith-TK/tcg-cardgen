import React from 'react'

function LayerPanel({ layers, selectedLayer, onSelectLayer, onUpdateLayer, onDeleteLayer }) {
  const getLayerIcon = (type) => {
    switch (type) {
      case 'text': return '📝'
      case 'shape': return '🟦'
      case 'image': return '🖼️'
      default: return '📄'
    }
  }

  const toggleVisibility = (layer) => {
    onUpdateLayer(layer.id, { visible: !layer.visible })
  }

  const duplicateLayer = (layer) => {
    const duplicated = {
      ...layer,
      id: `${layer.id}_copy_${Date.now()}`,
      name: `${layer.name} Copy`,
      region: {
        ...layer.region,
        x: layer.region.x + 10,
        y: layer.region.y + 10
      }
    }
    
    // Call parent to add the duplicated layer
    onAddLayer && onAddLayer(duplicated)
  }

  return (
    <div className="panel layer-panel" style={{ width: '250px' }}>
      <div className="panel-header">
        <h3>📑 Layers</h3>
      </div>
      
      <div className="panel-content">
        {layers.length === 0 ? (
          <p className="empty-state">No layers yet. Add a layer to get started.</p>
        ) : (
          <div className="layer-list">
            {layers.slice().reverse().map((layer) => (
              <div
                key={layer.id}
                className={`layer-item ${selectedLayer?.id === layer.id ? 'selected' : ''}`}
                onClick={() => onSelectLayer(layer)}
              >
                <div className="layer-main">
                  <span className="layer-icon">{getLayerIcon(layer.type)}</span>
                  <span className="layer-name">{layer.name}</span>
                </div>
                
                <div className="layer-controls">
                  <button
                    className={`layer-btn ${layer.visible ? 'active' : ''}`}
                    onClick={(e) => {
                      e.stopPropagation()
                      toggleVisibility(layer)
                    }}
                    title="Toggle visibility"
                  >
                    {layer.visible ? '👁️' : '🙈'}
                  </button>
                  
                  <button
                    className="layer-btn"
                    onClick={(e) => {
                      e.stopPropagation()
                      duplicateLayer(layer)
                    }}
                    title="Duplicate layer"
                  >
                    📋
                  </button>
                  
                  <button
                    className="layer-btn danger"
                    onClick={(e) => {
                      e.stopPropagation()
                      if (confirm(`Delete layer "${layer.name}"?`)) {
                        onDeleteLayer(layer.id)
                      }
                    }}
                    title="Delete layer"
                  >
                    🗑️
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default LayerPanel