    import React, { useEffect, useRef, useState } from 'react'
import { Canvas, Rect, Circle, Textbox, FabricText, FabricImage } from 'fabric'

function CanvasEditor({ template, layers, selectedLayer, onSelectLayer, onUpdateLayer }) {
  const canvasRef = useRef(null)
  const fabricCanvas = useRef(null)
  const [zoom, setZoom] = useState(1)
  const objectMap = useRef(new Map()) // Map layer IDs to fabric objects
  const isUpdatingFromCanvas = useRef(false) // Flag to prevent rebuild during canvas updates

  useEffect(() => {
    if (canvasRef.current && !fabricCanvas.current) {
      try {
        // Initialize Fabric.js canvas
        fabricCanvas.current = new Canvas(canvasRef.current, {
          width: template.format.width,
          height: template.format.height,
          backgroundColor: '#ffffff',
          selection: true,
          preserveObjectStacking: true
        })

        // Handle object selection
        fabricCanvas.current.on('selection:created', (e) => {
          const obj = e.selected?.[0]
          if (obj && obj.layerId) {
            const layer = layers.find(l => l.id === obj.layerId)
            if (layer) {
              onSelectLayer(layer)
            }
          }
        })

        fabricCanvas.current.on('selection:updated', (e) => {
          const obj = e.selected?.[0]
          if (obj && obj.layerId) {
            const layer = layers.find(l => l.id === obj.layerId)
            if (layer) {
              onSelectLayer(layer)
            }
          }
        })

        fabricCanvas.current.on('selection:cleared', () => {
          // Don't clear selection automatically - let the user manage it
        })

        // Handle object modifications
        fabricCanvas.current.on('object:modified', (e) => {
          const obj = e.target
          if (obj && obj.layerId) {
            const newRegion = {
              x: Math.round(obj.left || 0),
              y: Math.round(obj.top || 0),
              width: Math.round((obj.width || 0) * (obj.scaleX || 1)),
              height: Math.round((obj.height || 0) * (obj.scaleY || 1))
            }
            
            // Reset scale and update object dimensions
            obj.set({
              width: newRegion.width,
              height: newRegion.height,
              scaleX: 1,
              scaleY: 1
            })
            fabricCanvas.current.renderAll()
            
            // Update the layer data without triggering canvas rebuild
            console.log('Canvas: updating layer', obj.layerId, 'to region', newRegion)
            onUpdateLayer(obj.layerId, { region: newRegion })
          }
        })

        // Handle canvas click to deselect
        fabricCanvas.current.on('mouse:down', (e) => {
          if (!e.target) {
            onSelectLayer(null)
          }
        })
        
        console.log('Fabric.js canvas initialized successfully')
      } catch (error) {
        console.error('Failed to initialize Fabric.js canvas:', error)
      }
    }

    return () => {
      if (fabricCanvas.current) {
        try {
          fabricCanvas.current.dispose()
          fabricCanvas.current = null
        } catch (error) {
          console.error('Error disposing canvas:', error)
        }
      }
    }
  }, [template.format])

  // Update canvas when layers change
  useEffect(() => {
    if (!fabricCanvas.current) return
    
    console.log('Canvas: rebuilding with', layers.length, 'layers')

    // Clear existing objects
    fabricCanvas.current.clear()
    objectMap.current.clear()

    // Add layers to canvas
    layers.forEach(layer => {
      if (!layer.visible) return

      let fabricObject = null

      switch (layer.type) {
        case 'text':
          fabricObject = new Textbox(layer.content || 'New Text', {
            left: layer.region.x,
            top: layer.region.y,
            width: layer.region.width,
            fontSize: layer.font?.size || 16,
            fontFamily: layer.font?.family || 'Arial',
            fontWeight: layer.font?.weight || 'normal',
            fill: layer.font?.color?.substring(0, 7) || '#000000',
            textAlign: layer.align || 'left'
          })
          break

        case 'shape':
          if (layer.shape === 'circle') {
            const radius = Math.min(layer.region.width, layer.region.height) / 2
            fabricObject = new Circle({
              left: layer.region.x,
              top: layer.region.y,
              radius: radius,
              fill: layer.fill?.substring(0, 7) || '#CCCCCC',
              stroke: layer.stroke?.substring(0, 7) || '#000000',
              strokeWidth: layer.stroke_width || 1
            })
          } else {
            fabricObject = new Rect({
              left: layer.region.x,
              top: layer.region.y,
              width: layer.region.width,
              height: layer.region.height,
              fill: layer.fill?.substring(0, 7) || '#CCCCCC',
              stroke: layer.stroke?.substring(0, 7) || '#000000',
              strokeWidth: layer.stroke_width || 1,
              rx: layer.corner_radius || 0,
              ry: layer.corner_radius || 0
            })
          }
          break

        case 'image':
          if (layer.src) {
            FabricImage.fromURL(layer.src).then((img) => {
              img.set({
                left: layer.region.x,
                top: layer.region.y,
                layerId: layer.id
              })
              img.scaleToWidth(layer.region.width)
              img.scaleToHeight(layer.region.height)
              fabricCanvas.current.add(img)
              objectMap.current.set(layer.id, img)
            }).catch(error => {
              console.error('Error loading image:', error)
            })
            return // Skip the normal add process for images
          } else {
            // Placeholder for image without src
            fabricObject = new Rect({
              left: layer.region.x,
              top: layer.region.y,
              width: layer.region.width,
              height: layer.region.height,
              fill: '#f0f0f0',
              stroke: '#cccccc',
              strokeWidth: 2,
              strokeDashArray: [5, 5]
            })
            
            // Add "Image" text
            const placeholderText = new FabricText('📷 Image', {
              left: layer.region.x + layer.region.width / 2,
              top: layer.region.y + layer.region.height / 2,
              fontSize: 16,
              fill: '#999999',
              textAlign: 'center',
              originX: 'center',
              originY: 'center',
              selectable: false,
              evented: false
            })
            
            fabricCanvas.current.add(fabricObject)
            fabricCanvas.current.add(placeholderText)
            
            fabricObject.layerId = layer.id
            objectMap.current.set(layer.id, fabricObject)
            return
          }
          break
      }

      if (fabricObject) {
        fabricObject.layerId = layer.id
        fabricCanvas.current.add(fabricObject)
        objectMap.current.set(layer.id, fabricObject)
      }
    })

    fabricCanvas.current.renderAll()
  }, [layers])

  // Update selection when selectedLayer changes
  useEffect(() => {
    if (!fabricCanvas.current) return

    fabricCanvas.current.discardActiveObject()
    
    if (selectedLayer) {
      const fabricObject = objectMap.current.get(selectedLayer.id)
      if (fabricObject) {
        fabricCanvas.current.setActiveObject(fabricObject)
      }
    }
    
    fabricCanvas.current.renderAll()
  }, [selectedLayer])

  const handleZoomIn = () => {
    const newZoom = Math.min(zoom * 1.2, 3)
    setZoom(newZoom)
    if (fabricCanvas.current) {
      fabricCanvas.current.setZoom(newZoom)
    }
  }

  const handleZoomOut = () => {
    const newZoom = Math.max(zoom / 1.2, 0.1)
    setZoom(newZoom)
    if (fabricCanvas.current) {
      fabricCanvas.current.setZoom(newZoom)
    }
  }

  const handleFitToView = () => {
    if (!fabricCanvas.current) return
    
    const container = canvasRef.current.parentElement
    const containerWidth = container.clientWidth - 40 // padding
    const containerHeight = container.clientHeight - 100 // toolbar space
    
    const scaleX = containerWidth / template.format.width
    const scaleY = containerHeight / template.format.height
    const newZoom = Math.min(scaleX, scaleY, 1)
    
    setZoom(newZoom)
    fabricCanvas.current.setZoom(newZoom)
  }

  return (
    <div className="canvas-editor">
      {/* Canvas Toolbar */}
      <div className="canvas-toolbar">
        <button className="btn btn-small btn-secondary" onClick={handleZoomOut}>
          🔍−
        </button>
        <span className="zoom-level">
          {Math.round(zoom * 100)}%
        </span>
        <button className="btn btn-small btn-secondary" onClick={handleZoomIn}>
          🔍+
        </button>
        <button className="btn btn-small btn-secondary" onClick={handleFitToView}>
          ⏹️ Fit
        </button>
        
        <div className="canvas-info">
          Canvas: {template.format.width} × {template.format.height}px
        </div>
      </div>

      {/* Canvas Container */}
      <div className="canvas-container">
        <div className="canvas-wrapper">
          <canvas ref={canvasRef} />
        </div>
      </div>
    </div>
  )
}

export default CanvasEditor