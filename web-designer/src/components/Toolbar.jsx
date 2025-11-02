import React from 'react'

function Toolbar({ onAddLayer, onSave, onExport }) {
  return (
    <div className="toolbar">
      <div className="toolbar-group">
        <button 
          className="btn btn-primary btn-small"
          onClick={() => onAddLayer('text')}
          title="Add Text Layer"
        >
          📝 Text
        </button>
        <button 
          className="btn btn-primary btn-small"
          onClick={() => onAddLayer('shape')}
          title="Add Shape Layer"
        >
          🟦 Shape
        </button>
        <button 
          className="btn btn-primary btn-small"
          onClick={() => onAddLayer('image')}
          title="Add Image Layer"
        >
          🖼️ Image
        </button>
      </div>
      
      <div className="toolbar-group">
        <button 
          className="btn btn-secondary btn-small"
          onClick={onSave}
          title="Save Template"
        >
          💾 Save
        </button>
        <button 
          className="btn btn-success btn-small"
          onClick={onExport}
          title="Export Template"
        >
          📦 Export
        </button>
      </div>
    </div>
  )
}

export default Toolbar