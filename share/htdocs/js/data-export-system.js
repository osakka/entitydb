/**
 * EntityDB Data Export System
 * Comprehensive export functionality with multiple formats and options
 */

class DataExportSystem {
    constructor() {
        this.supportedFormats = {
            json: {
                name: 'JSON',
                extension: 'json',
                mimeType: 'application/json',
                description: 'JavaScript Object Notation - standard data interchange format'
            },
            csv: {
                name: 'CSV',
                extension: 'csv',
                mimeType: 'text/csv',
                description: 'Comma-separated values - spreadsheet compatible format'
            },
            xml: {
                name: 'XML',
                extension: 'xml',
                mimeType: 'application/xml',
                description: 'Extensible Markup Language - structured data format'
            },
            yaml: {
                name: 'YAML',
                extension: 'yaml',
                mimeType: 'text/yaml',
                description: 'Human-readable data serialization standard'
            },
            txt: {
                name: 'Text',
                extension: 'txt',
                mimeType: 'text/plain',
                description: 'Plain text format - human readable'
            },
            zip: {
                name: 'ZIP Archive',
                extension: 'zip',
                mimeType: 'application/zip',
                description: 'Compressed archive with multiple files'
            }
        };

        this.exportOptions = {
            includeContent: true,
            includeMetadata: true,
            includeTimestamps: false,
            prettifyOutput: true,
            compressOutput: false,
            splitLargeFiles: false,
            maxFileSize: 10 * 1024 * 1024, // 10MB
            dateFormat: 'iso', // iso, timestamp, human
            contentEncoding: 'base64' // base64, text, binary
        };

        this.init();
    }

    init() {
        // Load user preferences
        this.loadPreferences();
    }

    // Main export functions
    async exportEntities(entities, format = 'json', options = {}) {
        const exportConfig = { ...this.exportOptions, ...options };
        
        try {
            // Prepare entities for export
            const processedEntities = this.preprocessEntities(entities, exportConfig);
            
            // Generate export data based on format
            let exportData;
            switch (format) {
                case 'json':
                    exportData = this.exportToJSON(processedEntities, exportConfig);
                    break;
                case 'csv':
                    exportData = this.exportToCSV(processedEntities, exportConfig);
                    break;
                case 'xml':
                    exportData = this.exportToXML(processedEntities, exportConfig);
                    break;
                case 'yaml':
                    exportData = this.exportToYAML(processedEntities, exportConfig);
                    break;
                case 'txt':
                    exportData = this.exportToText(processedEntities, exportConfig);
                    break;
                case 'zip':
                    exportData = await this.exportToZip(processedEntities, exportConfig);
                    break;
                default:
                    throw new Error(`Unsupported export format: ${format}`);
            }

            // Create download
            await this.downloadData(exportData, format, exportConfig);
            
            // Track export
            this.trackExport(entities.length, format);
            
            return true;
        } catch (error) {
            console.error('Export failed:', error);
            this.showNotification(`Export failed: ${error.message}`, 'error');
            return false;
        }
    }

    preprocessEntities(entities, options) {
        return entities.map(entity => {
            const processed = {};

            // Always include ID
            processed.id = entity.id;

            // Include metadata if requested
            if (options.includeMetadata) {
                if (entity.created_at) {
                    processed.created_at = this.formatDate(entity.created_at, options.dateFormat);
                }
                if (entity.updated_at) {
                    processed.updated_at = this.formatDate(entity.updated_at, options.dateFormat);
                }
            }

            // Process tags
            if (entity.tags && entity.tags.length > 0) {
                processed.tags = entity.tags.map(tag => {
                    if (options.includeTimestamps) {
                        return tag; // Keep original format with timestamps
                    } else {
                        return this.stripTimestamp(tag);
                    }
                });
            }

            // Process content
            if (options.includeContent && entity.content) {
                processed.content = this.processContent(entity.content, options);
            }

            // Add search score if available
            if (entity._searchScore !== undefined) {
                processed.search_score = entity._searchScore;
            }

            return processed;
        });
    }

    processContent(content, options) {
        if (!content) return null;

        switch (options.contentEncoding) {
            case 'text':
                try {
                    return atob(content);
                } catch (e) {
                    return '[Binary content - cannot decode as text]';
                }
            case 'base64':
                return content;
            case 'binary':
                // For binary data, we'll include a reference or hash
                return {
                    type: 'binary',
                    size: content.length,
                    hash: this.generateHash(content)
                };
            default:
                return content;
        }
    }

    // Format-specific export functions
    exportToJSON(entities, options) {
        const exportObj = {
            metadata: this.generateExportMetadata(entities.length, 'json'),
            entities: entities
        };

        if (options.prettifyOutput) {
            return JSON.stringify(exportObj, null, 2);
        } else {
            return JSON.stringify(exportObj);
        }
    }

    exportToCSV(entities, options) {
        if (entities.length === 0) {
            return 'id,tags,content,created_at,updated_at\n';
        }

        // Determine columns based on first entity and options
        const columns = ['id'];
        
        if (options.includeMetadata) {
            columns.push('created_at', 'updated_at');
        }
        
        columns.push('tags');
        
        if (options.includeContent) {
            columns.push('content');
        }

        // Create CSV header
        let csv = columns.join(',') + '\n';

        // Add data rows
        entities.forEach(entity => {
            const row = columns.map(col => {
                let value = entity[col];
                
                if (value === undefined || value === null) {
                    return '';
                }
                
                if (Array.isArray(value)) {
                    value = value.join('; ');
                }
                
                if (typeof value === 'object') {
                    value = JSON.stringify(value);
                }
                
                // Escape CSV special characters
                value = String(value).replace(/"/g, '""');
                
                // Quote if contains comma, newline, or quote
                if (value.includes(',') || value.includes('\n') || value.includes('"')) {
                    value = `"${value}"`;
                }
                
                return value;
            });
            
            csv += row.join(',') + '\n';
        });

        return csv;
    }

    exportToXML(entities, options) {
        let xml = '<?xml version="1.0" encoding="UTF-8"?>\n';
        xml += '<entitydb_export>\n';
        xml += `  <metadata>\n`;
        xml += `    <export_date>${new Date().toISOString()}</export_date>\n`;
        xml += `    <entity_count>${entities.length}</entity_count>\n`;
        xml += `    <format>xml</format>\n`;
        xml += `  </metadata>\n`;
        xml += '  <entities>\n';

        entities.forEach(entity => {
            xml += '    <entity>\n';
            xml += `      <id>${this.escapeXml(entity.id)}</id>\n`;
            
            if (entity.created_at) {
                xml += `      <created_at>${this.escapeXml(entity.created_at)}</created_at>\n`;
            }
            
            if (entity.updated_at) {
                xml += `      <updated_at>${this.escapeXml(entity.updated_at)}</updated_at>\n`;
            }
            
            if (entity.tags && entity.tags.length > 0) {
                xml += '      <tags>\n';
                entity.tags.forEach(tag => {
                    xml += `        <tag>${this.escapeXml(tag)}</tag>\n`;
                });
                xml += '      </tags>\n';
            }
            
            if (entity.content) {
                xml += `      <content>${this.escapeXml(entity.content)}</content>\n`;
            }
            
            xml += '    </entity>\n';
        });

        xml += '  </entities>\n';
        xml += '</entitydb_export>\n';
        
        return xml;
    }

    exportToYAML(entities, options) {
        // Simple YAML implementation
        let yaml = '# EntityDB Export\n';
        yaml += `export_date: "${new Date().toISOString()}"\n`;
        yaml += `entity_count: ${entities.length}\n`;
        yaml += `format: yaml\n\n`;
        yaml += 'entities:\n';

        entities.forEach(entity => {
            yaml += `  - id: "${entity.id}"\n`;
            
            if (entity.created_at) {
                yaml += `    created_at: "${entity.created_at}"\n`;
            }
            
            if (entity.updated_at) {
                yaml += `    updated_at: "${entity.updated_at}"\n`;
            }
            
            if (entity.tags && entity.tags.length > 0) {
                yaml += '    tags:\n';
                entity.tags.forEach(tag => {
                    yaml += `      - "${tag}"\n`;
                });
            }
            
            if (entity.content) {
                yaml += `    content: "${entity.content}"\n`;
            }
            
            yaml += '\n';
        });

        return yaml;
    }

    exportToText(entities, options) {
        let text = 'EntityDB Export\n';
        text += '===============\n\n';
        text += `Export Date: ${new Date().toISOString()}\n`;
        text += `Entity Count: ${entities.length}\n`;
        text += `Format: Plain Text\n\n`;

        entities.forEach((entity, index) => {
            text += `Entity ${index + 1}: ${entity.id}\n`;
            text += '-'.repeat(50) + '\n';
            
            if (entity.created_at) {
                text += `Created: ${entity.created_at}\n`;
            }
            
            if (entity.updated_at) {
                text += `Updated: ${entity.updated_at}\n`;
            }
            
            if (entity.tags && entity.tags.length > 0) {
                text += `Tags: ${entity.tags.join(', ')}\n`;
            }
            
            if (entity.content) {
                text += `Content: ${entity.content}\n`;
            }
            
            text += '\n';
        });

        return text;
    }

    async exportToZip(entities, options) {
        // This would require a ZIP library like JSZip
        // For now, we'll create a JSON export with metadata
        const files = {};
        
        // Create manifest
        files['manifest.json'] = JSON.stringify({
            export_date: new Date().toISOString(),
            entity_count: entities.length,
            format: 'zip_archive',
            files: ['entities.json', 'metadata.json']
        }, null, 2);

        // Create entities file
        files['entities.json'] = this.exportToJSON(entities, options);
        
        // Create metadata file
        files['metadata.json'] = JSON.stringify(
            this.generateExportMetadata(entities.length, 'zip'), 
            null, 
            2
        );

        // For now, return as JSON (would need ZIP library for actual ZIP)
        return JSON.stringify(files, null, 2);
    }

    // Utility functions
    generateExportMetadata(entityCount, format) {
        return {
            export_date: new Date().toISOString(),
            entity_count: entityCount,
            format: format,
            version: '1.0',
            exported_by: 'EntityDB Web Interface',
            options: this.exportOptions
        };
    }

    formatDate(timestamp, format) {
        if (!timestamp) return null;
        
        const date = new Date(timestamp / 1000000); // Convert from nanoseconds
        
        switch (format) {
            case 'iso':
                return date.toISOString();
            case 'timestamp':
                return timestamp;
            case 'human':
                return date.toLocaleString();
            default:
                return date.toISOString();
        }
    }

    stripTimestamp(tag) {
        if (typeof tag !== 'string') return tag;
        const pipeIndex = tag.indexOf('|');
        return pipeIndex !== -1 ? tag.substring(pipeIndex + 1) : tag;
    }

    escapeXml(text) {
        if (typeof text !== 'string') return text;
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&apos;');
    }

    generateHash(content) {
        // Simple hash function for content identification
        let hash = 0;
        for (let i = 0; i < content.length; i++) {
            const char = content.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash = hash & hash; // Convert to 32bit integer
        }
        return hash.toString(16);
    }

    async downloadData(data, format, options) {
        const formatConfig = this.supportedFormats[format];
        if (!formatConfig) {
            throw new Error(`Unknown format: ${format}`);
        }

        // Create blob
        const blob = new Blob([data], { type: formatConfig.mimeType });
        
        // Generate filename
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const filename = `entitydb_export_${timestamp}.${formatConfig.extension}`;
        
        // Create download link
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        
        // Trigger download
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        // Cleanup
        setTimeout(() => URL.revokeObjectURL(url), 1000);
        
        this.showNotification(`Export downloaded: ${filename}`, 'success');
    }

    // Export dialog functions
    showExportDialog(entities) {
        if (!entities || entities.length === 0) {
            this.showNotification('No entities to export', 'warning');
            return;
        }

        const dialog = this.createExportDialog(entities);
        this.showModal(dialog);
    }

    createExportDialog(entities) {
        const modalId = 'export-dialog';
        const modalContent = `
            <div class="export-dialog">
                <div class="export-summary">
                    <h3>Export ${entities.length} Entities</h3>
                    <p class="text-muted">Choose format and options for your data export</p>
                </div>

                <form id="export-form" class="export-form">
                    <!-- Format Selection -->
                    <div class="form-section">
                        <h4 class="section-title">Export Format</h4>
                        <div class="format-grid">
                            ${Object.entries(this.supportedFormats).map(([key, format]) => `
                                <label class="format-option">
                                    <input type="radio" name="format" value="${key}" ${key === 'json' ? 'checked' : ''}>
                                    <div class="format-card">
                                        <strong>${format.name}</strong>
                                        <small>${format.description}</small>
                                    </div>
                                </label>
                            `).join('')}
                        </div>
                    </div>

                    <!-- Export Options -->
                    <div class="form-section">
                        <h4 class="section-title">Export Options</h4>
                        <div class="options-grid">
                            <label class="option-item">
                                <input type="checkbox" name="includeContent" checked>
                                <span>Include Content</span>
                            </label>
                            <label class="option-item">
                                <input type="checkbox" name="includeMetadata" checked>
                                <span>Include Metadata</span>
                            </label>
                            <label class="option-item">
                                <input type="checkbox" name="includeTimestamps">
                                <span>Include Timestamps</span>
                            </label>
                            <label class="option-item">
                                <input type="checkbox" name="prettifyOutput" checked>
                                <span>Prettify Output</span>
                            </label>
                        </div>
                    </div>

                    <!-- Advanced Options -->
                    <div class="form-section collapsible">
                        <h4 class="section-title clickable" onclick="this.parentElement.classList.toggle('expanded')">
                            <i class="fas fa-chevron-right expand-icon"></i>
                            Advanced Options
                        </h4>
                        <div class="section-content">
                            <div class="form-row">
                                <div class="form-group">
                                    <label class="form-label">Date Format</label>
                                    <select name="dateFormat" class="form-input">
                                        <option value="iso">ISO 8601</option>
                                        <option value="timestamp">Timestamp</option>
                                        <option value="human">Human Readable</option>
                                    </select>
                                </div>
                                <div class="form-group">
                                    <label class="form-label">Content Encoding</label>
                                    <select name="contentEncoding" class="form-input">
                                        <option value="base64">Base64</option>
                                        <option value="text">Text</option>
                                        <option value="binary">Binary Info</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>
                </form>

                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="dataExportSystem.closeModal('${modalId}')">
                        Cancel
                    </button>
                    <button type="button" class="btn btn-primary" onclick="dataExportSystem.performExport('${modalId}')">
                        <i class="fas fa-download"></i> Export Data
                    </button>
                </div>
            </div>
        `;

        return this.createModal(modalId, modalContent, {
            title: 'Export Entities',
            size: 'large'
        });
    }

    async performExport(modalId) {
        const form = document.getElementById('export-form');
        if (!form) return;

        const formData = new FormData(form);
        const format = formData.get('format') || 'json';
        
        const options = {
            includeContent: formData.has('includeContent'),
            includeMetadata: formData.has('includeMetadata'),
            includeTimestamps: formData.has('includeTimestamps'),
            prettifyOutput: formData.has('prettifyOutput'),
            dateFormat: formData.get('dateFormat') || 'iso',
            contentEncoding: formData.get('contentEncoding') || 'base64'
        };

        // Get entities to export (this would be passed to the dialog)
        const entities = this.currentExportEntities || [];
        
        try {
            await this.exportEntities(entities, format, options);
            this.closeModal(modalId);
        } catch (error) {
            console.error('Export failed:', error);
        }
    }

    // Modal system integration
    createModal(id, content, options = {}) {
        const modal = document.createElement('div');
        modal.id = id;
        modal.className = `modal ${options.size || 'medium'}`;
        modal.innerHTML = `
            <div class="modal-backdrop" onclick="dataExportSystem.closeModal('${id}')"></div>
            <div class="modal-dialog">
                <div class="modal-header">
                    <h2 class="modal-title">${options.title || 'Export'}</h2>
                    <button class="modal-close" onclick="dataExportSystem.closeModal('${id}')">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    ${content}
                </div>
            </div>
        `;

        // Add to modal container
        const container = document.getElementById('entity-modal-container') || document.body;
        container.appendChild(modal);
        
        return modal;
    }

    showModal(modal) {
        modal.classList.add('show');
        document.body.classList.add('modal-open');
    }

    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.remove();
                document.body.classList.remove('modal-open');
            }, 300);
        }
    }

    // Public API
    exportEntity(entity, format = 'json') {
        return this.exportEntities([entity], format);
    }

    exportSelectedEntities(entities, format = 'json') {
        this.currentExportEntities = entities;
        this.showExportDialog(entities);
    }

    getAvailableFormats() {
        return Object.keys(this.supportedFormats);
    }

    // Preferences management
    savePreferences() {
        localStorage.setItem('entitydb-export-options', JSON.stringify(this.exportOptions));
    }

    loadPreferences() {
        try {
            const saved = localStorage.getItem('entitydb-export-options');
            if (saved) {
                this.exportOptions = { ...this.exportOptions, ...JSON.parse(saved) };
            }
        } catch (e) {
            console.warn('Failed to load export preferences:', e);
        }
    }

    // Analytics
    trackExport(entityCount, format) {
        try {
            const stats = JSON.parse(localStorage.getItem('entitydb-export-stats') || '{}');
            stats[format] = (stats[format] || 0) + 1;
            stats.totalEntities = (stats.totalEntities || 0) + entityCount;
            localStorage.setItem('entitydb-export-stats', JSON.stringify(stats));
        } catch (e) {
            console.warn('Failed to track export:', e);
        }
    }

    getExportStats() {
        try {
            return JSON.parse(localStorage.getItem('entitydb-export-stats') || '{}');
        } catch (e) {
            return {};
        }
    }

    // Notification helper
    showNotification(message, type = 'info') {
        if (window.notificationSystem) {
            window.notificationSystem.show(message, type);
        } else {
            console.log(`${type}: ${message}`);
        }
    }
}

// Initialize the export system
if (typeof window !== 'undefined') {
    window.DataExportSystem = DataExportSystem;
    
    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.dataExportSystem = new DataExportSystem();
        });
    } else {
        window.dataExportSystem = new DataExportSystem();
    }
}