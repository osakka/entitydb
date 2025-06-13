/**
 * EntityDB Modal System for Entity Management
 * Complete modal implementation for create, edit, and view operations
 */

class EntityModalSystem {
    constructor() {
        this.activeModal = null;
        this.validators = {};
        this.setupModalContainer();
    }

    setupModalContainer() {
        if (!document.getElementById('entity-modal-container')) {
            const container = document.createElement('div');
            container.id = 'entity-modal-container';
            container.className = 'modal-container';
            document.body.appendChild(container);
        }
    }

    showEntityModal(mode, entity = null, options = {}) {
        const modalId = `entity-modal-${mode}`;
        this.closeAllModals();

        let modalContent;
        switch (mode) {
            case 'create':
                modalContent = this.createEntityModalContent(entity, options);
                break;
            case 'edit':
                modalContent = this.editEntityModalContent(entity, options);
                break;
            case 'view':
                modalContent = this.viewEntityModalContent(entity, options);
                break;
            default:
                console.error('Unknown modal mode:', mode);
                return;
        }

        const modal = this.createModal(modalId, modalContent, {
            title: this.getModalTitle(mode, entity),
            size: options.size || 'large',
            closable: options.closable !== false
        });

        this.showModal(modal);
        this.setupModalEventListeners(modal, mode, entity);
        
        // Focus first input
        const firstInput = modal.querySelector('input, textarea, select');
        if (firstInput) {
            setTimeout(() => firstInput.focus(), 100);
        }

        return modal;
    }

    getModalTitle(mode, entity) {
        switch (mode) {
            case 'create':
                return 'Create New Entity';
            case 'edit':
                return `Edit Entity: ${this.getEntityDisplayName(entity)}`;
            case 'view':
                return `View Entity: ${this.getEntityDisplayName(entity)}`;
            default:
                return 'Entity';
        }
    }

    getEntityDisplayName(entity) {
        if (!entity) return 'Unknown';
        
        if (entity.tags) {
            for (const tag of entity.tags) {
                const cleanTag = this.stripTimestamp(tag);
                if (cleanTag.startsWith('title:') || cleanTag.startsWith('name:')) {
                    return cleanTag.split(':').slice(1).join(':') || entity.id.substring(0, 8);
                }
            }
        }
        
        return entity.id.substring(0, 8);
    }

    createEntityModalContent(entity, options) {
        return `
            <form id="entity-form" class="entity-form">
                <div class="form-sections">
                    <!-- Basic Information -->
                    <div class="form-section">
                        <h3 class="section-title">Basic Information</h3>
                        <div class="form-row">
                            <div class="form-group">
                                <label class="form-label" for="entity-title">Title</label>
                                <input 
                                    type="text" 
                                    id="entity-title" 
                                    name="title"
                                    class="form-input" 
                                    placeholder="Enter entity title"
                                    required
                                >
                                <small class="form-hint">A human-readable title for this entity</small>
                            </div>
                        </div>
                        <div class="form-row">
                            <div class="form-group half-width">
                                <label class="form-label" for="entity-type">Type</label>
                                <select id="entity-type" name="type" class="form-input">
                                    <option value="">Select type...</option>
                                    <option value="document">Document</option>
                                    <option value="user">User</option>
                                    <option value="task">Task</option>
                                    <option value="note">Note</option>
                                    <option value="file">File</option>
                                    <option value="custom">Custom</option>
                                </select>
                            </div>
                            <div class="form-group half-width">
                                <label class="form-label" for="entity-status">Status</label>
                                <select id="entity-status" name="status" class="form-input">
                                    <option value="">Select status...</option>
                                    <option value="active">Active</option>
                                    <option value="inactive">Inactive</option>
                                    <option value="draft">Draft</option>
                                    <option value="archived">Archived</option>
                                </select>
                            </div>
                        </div>
                    </div>

                    <!-- Tags -->
                    <div class="form-section">
                        <h3 class="section-title">Tags</h3>
                        <div class="form-group">
                            <label class="form-label" for="entity-tags">Tags</label>
                            <div class="tag-input-container">
                                <input 
                                    type="text" 
                                    id="entity-tags-input" 
                                    class="form-input tag-input" 
                                    placeholder="Type a tag and press Enter"
                                >
                                <div class="tag-suggestions" id="tag-suggestions"></div>
                            </div>
                            <div class="tags-display" id="tags-display"></div>
                            <small class="form-hint">Press Enter to add tags. Use format "key:value" for structured tags.</small>
                        </div>
                    </div>

                    <!-- Content -->
                    <div class="form-section">
                        <h3 class="section-title">Content</h3>
                        <div class="form-group">
                            <div class="content-type-selector">
                                <label class="radio-label">
                                    <input type="radio" name="content-type" value="text" checked>
                                    <span>Text Content</span>
                                </label>
                                <label class="radio-label">
                                    <input type="radio" name="content-type" value="file">
                                    <span>File Upload</span>
                                </label>
                                <label class="radio-label">
                                    <input type="radio" name="content-type" value="json">
                                    <span>JSON Data</span>
                                </label>
                            </div>
                        </div>
                        
                        <div class="content-input text-content">
                            <label class="form-label" for="entity-content-text">Text Content</label>
                            <textarea 
                                id="entity-content-text" 
                                name="content-text"
                                class="form-input content-textarea" 
                                rows="6"
                                placeholder="Enter text content..."
                            ></textarea>
                        </div>

                        <div class="content-input file-content" style="display: none;">
                            <label class="form-label" for="entity-content-file">File Upload</label>
                            <div class="file-upload-area" id="file-upload-area">
                                <input type="file" id="entity-content-file" name="content-file" class="file-input">
                                <div class="file-upload-placeholder">
                                    <i class="fas fa-cloud-upload-alt"></i>
                                    <p>Click to upload or drag and drop</p>
                                    <small>Maximum file size: 10MB</small>
                                </div>
                                <div class="file-preview" id="file-preview" style="display: none;"></div>
                            </div>
                        </div>

                        <div class="content-input json-content" style="display: none;">
                            <label class="form-label" for="entity-content-json">JSON Data</label>
                            <textarea 
                                id="entity-content-json" 
                                name="content-json"
                                class="form-input content-textarea json-editor" 
                                rows="8"
                                placeholder='{"key": "value"}'
                            ></textarea>
                            <div class="json-validation" id="json-validation"></div>
                        </div>
                    </div>

                    <!-- Advanced Options -->
                    <div class="form-section collapsible">
                        <h3 class="section-title clickable" onclick="this.parentElement.classList.toggle('expanded')">
                            <i class="fas fa-chevron-right expand-icon"></i>
                            Advanced Options
                        </h3>
                        <div class="section-content">
                            <div class="form-row">
                                <div class="form-group half-width">
                                    <label class="form-label" for="entity-dataset">Dataset</label>
                                    <select id="entity-dataset" name="dataset" class="form-input">
                                        <option value="default">Default</option>
                                        <option value="_system">System</option>
                                    </select>
                                </div>
                                <div class="form-group half-width">
                                    <label class="form-label" for="entity-permissions">Permissions</label>
                                    <select id="entity-permissions" name="permissions" class="form-input">
                                        <option value="public">Public</option>
                                        <option value="private">Private</option>
                                        <option value="restricted">Restricted</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="entityModalSystem.closeAllModals()">
                        Cancel
                    </button>
                    <button type="submit" class="btn btn-primary">
                        <i class="fas fa-save"></i> Create Entity
                    </button>
                </div>
            </form>
        `;
    }

    editEntityModalContent(entity, options) {
        const content = this.createEntityModalContent(entity, options);
        
        // Modify for edit mode
        return content
            .replace('Create New Entity', `Edit Entity: ${this.getEntityDisplayName(entity)}`)
            .replace('Create Entity', 'Save Changes')
            .replace('Enter entity title', entity ? this.getEntityTitle(entity) : 'Enter entity title');
    }

    viewEntityModalContent(entity, options) {
        return `
            <div class="entity-view">
                <div class="entity-view-header">
                    <div class="entity-info">
                        <h2 class="entity-title">${this.escapeHtml(this.getEntityTitle(entity))}</h2>
                        <div class="entity-meta">
                            <span class="meta-item">
                                <i class="fas fa-fingerprint"></i>
                                <strong>ID:</strong> <code>${entity.id}</code>
                            </span>
                            <span class="meta-item">
                                <i class="fas fa-clock"></i>
                                <strong>Created:</strong> ${this.formatDate(entity.created_at)}
                            </span>
                            <span class="meta-item">
                                <i class="fas fa-edit"></i>
                                <strong>Updated:</strong> ${this.formatDate(entity.updated_at)}
                            </span>
                        </div>
                    </div>
                    <div class="entity-actions">
                        <button class="btn btn-secondary" onclick="entityBrowserEnhanced.editEntity('${entity.id}'); entityModalSystem.closeAllModals();">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="btn btn-outline-secondary" onclick="entityModalSystem.exportEntity('${entity.id}')">
                            <i class="fas fa-download"></i> Export
                        </button>
                    </div>
                </div>

                <div class="entity-view-content">
                    <div class="view-section">
                        <h3 class="section-title">Tags</h3>
                        <div class="tags-display">
                            ${this.renderEntityTags(entity)}
                        </div>
                    </div>

                    <div class="view-section">
                        <h3 class="section-title">Content</h3>
                        <div class="content-display">
                            ${this.renderEntityContent(entity)}
                        </div>
                    </div>

                    ${entity.relationships ? `
                        <div class="view-section">
                            <h3 class="section-title">Relationships</h3>
                            <div class="relationships-display">
                                ${this.renderEntityRelationships(entity.relationships)}
                            </div>
                        </div>
                    ` : ''}
                </div>

                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="entityModalSystem.closeAllModals()">
                        Close
                    </button>
                </div>
            </div>
        `;
    }

    createModal(id, content, options = {}) {
        const modal = document.createElement('div');
        modal.id = id;
        modal.className = `modal ${options.size || 'medium'}`;
        modal.innerHTML = `
            <div class="modal-backdrop" onclick="entityModalSystem.closeModal('${id}')"></div>
            <div class="modal-dialog">
                <div class="modal-header">
                    <h2 class="modal-title">${options.title || 'Modal'}</h2>
                    ${options.closable !== false ? `
                        <button class="modal-close" onclick="entityModalSystem.closeModal('${id}')">
                            <i class="fas fa-times"></i>
                        </button>
                    ` : ''}
                </div>
                <div class="modal-body">
                    ${content}
                </div>
            </div>
        `;

        document.getElementById('entity-modal-container').appendChild(modal);
        return modal;
    }

    showModal(modal) {
        this.activeModal = modal;
        modal.classList.add('show');
        document.body.classList.add('modal-open');
        
        // Trap focus within modal
        this.trapFocus(modal);
    }

    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.remove();
                if (this.activeModal === modal) {
                    this.activeModal = null;
                    document.body.classList.remove('modal-open');
                }
            }, 300);
        }
    }

    closeAllModals() {
        const modals = document.querySelectorAll('.modal.show');
        modals.forEach(modal => {
            this.closeModal(modal.id);
        });
    }

    setupModalEventListeners(modal, mode, entity) {
        if (mode === 'create' || mode === 'edit') {
            this.setupFormEventListeners(modal, mode, entity);
        }
        
        // Escape key to close
        const escapeHandler = (e) => {
            if (e.key === 'Escape') {
                this.closeModal(modal.id);
                document.removeEventListener('keydown', escapeHandler);
            }
        };
        document.addEventListener('keydown', escapeHandler);
    }

    setupFormEventListeners(modal, mode, entity) {
        const form = modal.querySelector('#entity-form');
        if (!form) return;

        // Pre-populate form for edit mode
        if (mode === 'edit' && entity) {
            this.populateForm(form, entity);
        }

        // Content type switching
        const contentTypeRadios = form.querySelectorAll('input[name="content-type"]');
        contentTypeRadios.forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.toggleContentType(e.target.value);
            });
        });

        // Tag input handling
        this.setupTagInput(form);

        // File upload handling
        this.setupFileUpload(form);

        // JSON validation
        this.setupJSONValidation(form);

        // Form submission
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.handleFormSubmission(form, mode, entity);
        });
    }

    setupTagInput(form) {
        const tagInput = form.querySelector('#entity-tags-input');
        const tagsDisplay = form.querySelector('#tags-display');
        let tags = [];

        if (!tagInput || !tagsDisplay) return;

        const renderTags = () => {
            tagsDisplay.innerHTML = tags.map((tag, index) => `
                <span class="tag-item">
                    ${this.escapeHtml(tag)}
                    <button type="button" class="tag-remove" onclick="this.parentElement.remove(); entityModalSystem.removeTag(${index})">
                        <i class="fas fa-times"></i>
                    </button>
                </span>
            `).join('');
        };

        const addTag = (tagText) => {
            const tag = tagText.trim();
            if (tag && !tags.includes(tag)) {
                tags.push(tag);
                renderTags();
                tagInput.value = '';
            }
        };

        tagInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addTag(tagInput.value);
            }
        });

        // Store tags array for later access
        form.entityTags = tags;
        window.removeTag = (index) => {
            tags.splice(index, 1);
            renderTags();
        };
    }

    setupFileUpload(form) {
        const fileInput = form.querySelector('#entity-content-file');
        const uploadArea = form.querySelector('#file-upload-area');
        const preview = form.querySelector('#file-preview');

        if (!fileInput || !uploadArea || !preview) return;

        // Drag and drop
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });

        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                fileInput.files = files;
                this.handleFileSelection(files[0], preview);
            }
        });

        uploadArea.addEventListener('click', () => {
            fileInput.click();
        });

        fileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                this.handleFileSelection(e.target.files[0], preview);
            }
        });
    }

    handleFileSelection(file, preview) {
        preview.style.display = 'block';
        preview.innerHTML = `
            <div class="file-info">
                <i class="fas fa-file"></i>
                <div class="file-details">
                    <strong>${this.escapeHtml(file.name)}</strong>
                    <small>${this.formatFileSize(file.size)} • ${file.type || 'Unknown type'}</small>
                </div>
            </div>
        `;
    }

    setupJSONValidation(form) {
        const jsonTextarea = form.querySelector('#entity-content-json');
        const validation = form.querySelector('#json-validation');

        if (!jsonTextarea || !validation) return;

        const validateJSON = () => {
            const value = jsonTextarea.value.trim();
            if (!value) {
                validation.innerHTML = '';
                return true;
            }

            try {
                JSON.parse(value);
                validation.innerHTML = '<span class="validation-success"><i class="fas fa-check"></i> Valid JSON</span>';
                return true;
            } catch (error) {
                validation.innerHTML = `<span class="validation-error"><i class="fas fa-exclamation-triangle"></i> ${error.message}</span>`;
                return false;
            }
        };

        jsonTextarea.addEventListener('input', validateJSON);
        jsonTextarea.addEventListener('blur', validateJSON);
    }

    toggleContentType(type) {
        const containers = document.querySelectorAll('.content-input');
        containers.forEach(container => {
            container.style.display = 'none';
        });

        const activeContainer = document.querySelector(`.${type}-content`);
        if (activeContainer) {
            activeContainer.style.display = 'block';
        }
    }

    async handleFormSubmission(form, mode, entity) {
        const submitButton = form.querySelector('button[type="submit"]');
        const originalText = submitButton.innerHTML;
        
        try {
            submitButton.disabled = true;
            submitButton.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';

            const formData = this.collectFormData(form);
            const isUpdate = mode === 'edit';
            
            if (isUpdate && entity) {
                formData.id = entity.id;
            }

            const success = await window.entityBrowserEnhanced.saveEntity(formData, isUpdate);
            
            if (success) {
                this.closeAllModals();
            }
        } catch (error) {
            console.error('Form submission error:', error);
            this.showNotification('Failed to save entity: ' + error.message, 'error');
        } finally {
            submitButton.disabled = false;
            submitButton.innerHTML = originalText;
        }
    }

    collectFormData(form) {
        const data = {
            tags: [],
            content: '',
            dataset: form.dataset?.value || 'default'
        };

        // Collect basic fields
        const title = form.title?.value?.trim();
        if (title) data.tags.push(`title:${title}`);
        
        const type = form.type?.value?.trim();
        if (type) data.tags.push(`type:${type}`);
        
        const status = form.status?.value?.trim();
        if (status) data.tags.push(`status:${status}`);

        // Add custom tags
        if (form.entityTags) {
            data.tags.push(...form.entityTags);
        }

        // Collect content based on type
        const contentType = form.querySelector('input[name="content-type"]:checked')?.value;
        switch (contentType) {
            case 'text':
                data.content = btoa(form['content-text']?.value || '');
                break;
            case 'file':
                // Handle file upload - convert to base64
                const file = form['content-file']?.files[0];
                if (file) {
                    return this.handleFileUpload(file).then(base64Content => {
                        data.content = base64Content;
                        return data;
                    });
                }
                break;
            case 'json':
                const jsonContent = form['content-json']?.value?.trim();
                if (jsonContent) {
                    try {
                        JSON.parse(jsonContent); // Validate
                        data.content = btoa(jsonContent);
                    } catch (e) {
                        throw new Error('Invalid JSON content');
                    }
                }
                break;
        }

        return Promise.resolve(data);
    }

    async handleFileUpload(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = () => {
                const base64 = reader.result.split(',')[1]; // Remove data URL prefix
                resolve(base64);
            };
            reader.onerror = reject;
            reader.readAsDataURL(file);
        });
    }

    populateForm(form, entity) {
        if (!entity || !entity.tags) return;

        // Populate basic fields from tags
        entity.tags.forEach(tag => {
            const cleanTag = this.stripTimestamp(tag);
            const [key, ...valueParts] = cleanTag.split(':');
            const value = valueParts.join(':');

            switch (key) {
                case 'title':
                    const titleField = form.querySelector('#entity-title');
                    if (titleField) titleField.value = value;
                    break;
                case 'type':
                    const typeField = form.querySelector('#entity-type');
                    if (typeField) typeField.value = value;
                    break;
                case 'status':
                    const statusField = form.querySelector('#entity-status');
                    if (statusField) statusField.value = value;
                    break;
            }
        });

        // Populate content
        if (entity.content) {
            try {
                const decodedContent = atob(entity.content);
                const textArea = form.querySelector('#entity-content-text');
                if (textArea) {
                    textArea.value = decodedContent;
                }
            } catch (e) {
                console.warn('Could not decode entity content as text');
            }
        }

        // Add existing tags to tag system
        const customTags = entity.tags
            .map(tag => this.stripTimestamp(tag))
            .filter(tag => !tag.startsWith('title:') && !tag.startsWith('type:') && !tag.startsWith('status:'));
        
        if (form.entityTags) {
            form.entityTags.push(...customTags);
            // Re-render tags
            const event = new Event('input');
            form.querySelector('#entity-tags-input')?.dispatchEvent(event);
        }
    }

    // Utility methods
    stripTimestamp(tag) {
        if (typeof tag !== 'string') return tag;
        const pipeIndex = tag.indexOf('|');
        return pipeIndex !== -1 ? tag.substring(pipeIndex + 1) : tag;
    }

    getEntityTitle(entity) {
        if (!entity?.tags) return 'Untitled Entity';
        
        for (const tag of entity.tags) {
            const cleanTag = this.stripTimestamp(tag);
            if (cleanTag.startsWith('title:')) {
                return cleanTag.split(':').slice(1).join(':') || 'Untitled Entity';
            }
        }
        
        return 'Untitled Entity';
    }

    renderEntityTags(entity) {
        if (!entity?.tags || entity.tags.length === 0) {
            return '<span class="text-muted">No tags</span>';
        }

        return entity.tags.map(tag => {
            const cleanTag = this.stripTimestamp(tag);
            return `<span class="badge badge-secondary">${this.escapeHtml(cleanTag)}</span>`;
        }).join(' ');
    }

    renderEntityContent(entity) {
        if (!entity.content) {
            return '<span class="text-muted">No content</span>';
        }

        try {
            const decoded = atob(entity.content);
            
            // Try to parse as JSON
            try {
                const json = JSON.parse(decoded);
                return `<pre class="json-content"><code>${this.escapeHtml(JSON.stringify(json, null, 2))}</code></pre>`;
            } catch (e) {
                // Not JSON, treat as text
                return `<div class="text-content">${this.escapeHtml(decoded)}</div>`;
            }
        } catch (e) {
            return '<span class="text-muted">Binary content (cannot display)</span>';
        }
    }

    renderEntityRelationships(relationships) {
        if (!relationships || relationships.length === 0) {
            return '<span class="text-muted">No relationships</span>';
        }

        return relationships.map(rel => `
            <div class="relationship-item">
                <strong>${rel.type}</strong>: ${rel.target_id}
            </div>
        `).join('');
    }

    formatDate(timestamp) {
        if (!timestamp) return 'Unknown';
        try {
            return new Date(timestamp / 1000000).toLocaleString();
        } catch (e) {
            return 'Invalid date';
        }
    }

    formatFileSize(bytes) {
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        if (bytes === 0) return '0 Bytes';
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    trapFocus(modal) {
        const focusableElements = modal.querySelectorAll(
            'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];

        modal.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                if (e.shiftKey) {
                    if (document.activeElement === firstElement) {
                        e.preventDefault();
                        lastElement.focus();
                    }
                } else {
                    if (document.activeElement === lastElement) {
                        e.preventDefault();
                        firstElement.focus();
                    }
                }
            }
        });
    }

    showNotification(message, type = 'info') {
        if (window.notificationSystem) {
            window.notificationSystem.show(message, type);
        } else {
            console.log(`${type}: ${message}`);
        }
    }

    exportEntity(entityId) {
        this.showNotification('Entity export - coming soon', 'info');
    }
}

// Initialize the modal system
if (typeof window !== 'undefined') {
    window.EntityModalSystem = EntityModalSystem;
    
    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.entityModalSystem = new EntityModalSystem();
        });
    } else {
        window.entityModalSystem = new EntityModalSystem();
    }
}