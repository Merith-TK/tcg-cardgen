import { useState, useEffect } from 'react'
import './App.css'
import CanvasEditor from './components/CanvasEditor'
import LayerPanel from './components/LayerPanel'
import PropertiesPanel from './components/PropertiesPanel'
import Toolbar from './components/Toolbar'

function App() {
  const [template, setTemplate] = useState({
    name: 'New Template',
    format: { width: 375, height: 525 },
    layers: []
  })
  const [selectedLayer, setSelectedLayer] = useState(null)
  const [layers, setLayers] = useState([])

  // Initialize with default template
  useEffect(() => {
    const defaultLayers = [
      {
        id: 'background',
        name: 'Background',
        type: 'shape',
        region: { x: 10, y: 10, width: 355, height: 505 },
        shape: 'rectangle',
        fill: '#F0F0F0FF',
        stroke: '#000000FF',
        stroke_width: 2,
        corner_radius: 10,
        visible: true
      },
      {
        id: 'title',
        name: 'Card Title',
        type: 'text',
        region: { x: 20, y: 20, width: 335, height: 40 },
        content: 'Card Title',
        font: {
          family: 'Arial',
          size: 24,
          weight: 'bold',
          color: '#000000FF'
        },
        align: 'center',
        visible: true
      }
    ]
    setLayers(defaultLayers)
    setTemplate(prev => ({ ...prev, layers: defaultLayers }))
  }, [])

  const addLayer = (type) => {
    const newLayer = {
      id: `layer_${Date.now()}`,
      name: `New ${type}`,
      type: type,
      region: { 
        x: 50 + Math.random() * 100, 
        y: 50 + Math.random() * 100, 
        width: 200, 
        height: type === 'text' ? 50 : 150 
      },
      visible: true
    }

    // Add type-specific properties
    switch (type) {
      case 'text':
        newLayer.content = 'New Text'
        newLayer.font = {
          family: 'Arial',
          size: 16,
          weight: 'normal',
          color: '#000000FF'
        }
        newLayer.align = 'left'
        break
      case 'shape':
        newLayer.shape = 'rectangle'
        newLayer.fill = '#CCCCCCFF'
        newLayer.stroke = '#000000FF'
        newLayer.stroke_width = 1
        break
      case 'image':
        newLayer.src = ''
        break
    }

    const updatedLayers = [...layers, newLayer]
    setLayers(updatedLayers)
    setTemplate(prev => ({ ...prev, layers: updatedLayers }))
    setSelectedLayer(newLayer)
  }

  const updateLayer = (layerId, updates) => {
    const updatedLayers = layers.map(layer => 
      layer.id === layerId ? { ...layer, ...updates } : layer
    )
    setLayers(updatedLayers)
    setTemplate(prev => ({ ...prev, layers: updatedLayers }))
    
    if (selectedLayer && selectedLayer.id === layerId) {
      setSelectedLayer({ ...selectedLayer, ...updates })
    }
  }

  const deleteLayer = (layerId) => {
    const updatedLayers = layers.filter(layer => layer.id !== layerId)
    setLayers(updatedLayers)
    setTemplate(prev => ({ ...prev, layers: updatedLayers }))
    
    if (selectedLayer && selectedLayer.id === layerId) {
      setSelectedLayer(null)
    }
  }

  const saveTemplate = async () => {
    try {
      const response = await fetch('http://localhost:3000/api/template/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(template)
      })
      
      if (response.ok) {
        alert('Template saved successfully!')
      } else {
        alert('Failed to save template')
      }
    } catch (error) {
      console.error('Error saving template:', error)
      alert('Error saving template')
    }
  }

  const exportTemplate = async () => {
    try {
      const response = await fetch('http://localhost:3000/api/render', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          template: template,
          card: {
            name: 'Sample Card',
            description: 'This is a sample card'
          }
        })
      })
      
      if (response.ok) {
        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `${template.name}.png`
        a.click()
        URL.revokeObjectURL(url)
      } else {
        alert('Failed to export template')
      }
    } catch (error) {
      console.error('Error exporting template:', error)
      alert('Error exporting template')
    }
  }

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-left">
          <h1>🎨 TCG Cardstyle Designer</h1>
          <span className="version">v2.0.0</span>
        </div>
        <div className="header-center">
          <input 
            type="text" 
            value={template.name}
            onChange={(e) => setTemplate(prev => ({ ...prev, name: e.target.value }))}
            className="template-name-input"
            placeholder="Template Name"
          />
        </div>
        <div className="header-right">
          <Toolbar 
            onAddLayer={addLayer}
            onSave={saveTemplate}
            onExport={exportTemplate}
          />
        </div>
      </header>

      <main className="app-main">
        <LayerPanel 
          layers={layers}
          selectedLayer={selectedLayer}
          onSelectLayer={setSelectedLayer}
          onUpdateLayer={updateLayer}
          onDeleteLayer={deleteLayer}
        />
        
        <CanvasEditor 
          template={template}
          layers={layers}
          selectedLayer={selectedLayer}
          onSelectLayer={setSelectedLayer}
          onUpdateLayer={updateLayer}
        />
        
        <PropertiesPanel 
          selectedLayer={selectedLayer}
          onUpdateLayer={updateLayer}
        />
      </main>
    </div>
  )
}

export default App
