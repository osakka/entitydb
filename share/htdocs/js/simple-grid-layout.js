// Simple Grid Layout Component for Vue 3
// A lightweight alternative to vue-grid-layout

const SimpleGridLayout = {
    name: 'SimpleGridLayout',
    props: {
        layout: {
            type: Array,
            default: () => []
        },
        colNum: {
            type: Number,
            default: 12
        },
        rowHeight: {
            type: Number,
            default: 30
        },
        margin: {
            type: Array,
            default: () => [10, 10]
        },
        isDraggable: {
            type: Boolean,
            default: true
        },
        isResizable: {
            type: Boolean,
            default: true
        }
    },
    emits: ['layout-updated'],
    template: `
        <div class="simple-grid-layout" ref="container">
            <slot></slot>
        </div>
    `,
    mounted() {
        this.initializeGrid();
    },
    methods: {
        initializeGrid() {
            // Calculate container width
            const container = this.$refs.container;
            if (!container) return;
            
            const containerWidth = container.offsetWidth;
            const colWidth = (containerWidth - (this.margin[0] * (this.colNum + 1))) / this.colNum;
            
            // Update CSS variable for children
            container.style.setProperty('--grid-col-width', colWidth + 'px');
            container.style.setProperty('--grid-row-height', this.rowHeight + 'px');
            container.style.setProperty('--grid-margin-x', this.margin[0] + 'px');
            container.style.setProperty('--grid-margin-y', this.margin[1] + 'px');
        }
    },
    watch: {
        layout: {
            deep: true,
            handler() {
                this.$nextTick(() => {
                    this.initializeGrid();
                });
            }
        }
    }
};

const SimpleGridItem = {
    name: 'SimpleGridItem',
    props: {
        x: {
            type: Number,
            default: 0
        },
        y: {
            type: Number,
            default: 0
        },
        w: {
            type: Number,
            default: 1
        },
        h: {
            type: Number,
            default: 1
        },
        i: {
            type: String,
            required: true
        }
    },
    emits: ['resize', 'move'],
    template: `
        <div 
            class="simple-grid-item" 
            :style="itemStyle"
            :data-grid-id="i"
        >
            <slot></slot>
            <div 
                v-if="$parent.isResizable" 
                class="resize-handle"
                @mousedown="startResize"
            ></div>
        </div>
    `,
    computed: {
        itemStyle() {
            const colWidth = parseFloat(getComputedStyle(this.$el?.parentElement || document.body).getPropertyValue('--grid-col-width') || 100);
            const rowHeight = parseFloat(getComputedStyle(this.$el?.parentElement || document.body).getPropertyValue('--grid-row-height') || 30);
            const marginX = parseFloat(getComputedStyle(this.$el?.parentElement || document.body).getPropertyValue('--grid-margin-x') || 10);
            const marginY = parseFloat(getComputedStyle(this.$el?.parentElement || document.body).getPropertyValue('--grid-margin-y') || 10);
            
            return {
                position: 'absolute',
                left: (this.x * (colWidth + marginX) + marginX) + 'px',
                top: (this.y * (rowHeight + marginY) + marginY) + 'px',
                width: (this.w * colWidth + (this.w - 1) * marginX) + 'px',
                height: (this.h * rowHeight + (this.h - 1) * marginY) + 'px',
                transition: 'all 200ms ease',
                userSelect: 'none'
            };
        }
    },
    mounted() {
        if (this.$parent.isDraggable) {
            this.makeDraggable();
        }
    },
    methods: {
        makeDraggable() {
            const el = this.$el;
            let startX, startY, startLeft, startTop;
            
            el.style.cursor = 'move';
            
            el.addEventListener('mousedown', (e) => {
                if (e.target.classList.contains('resize-handle')) return;
                
                startX = e.clientX;
                startY = e.clientY;
                startLeft = el.offsetLeft;
                startTop = el.offsetTop;
                
                el.classList.add('dragging');
                
                const handleMouseMove = (e) => {
                    e.preventDefault();
                    const dx = e.clientX - startX;
                    const dy = e.clientY - startY;
                    
                    el.style.left = (startLeft + dx) + 'px';
                    el.style.top = (startTop + dy) + 'px';
                };
                
                const handleMouseUp = () => {
                    el.classList.remove('dragging');
                    document.removeEventListener('mousemove', handleMouseMove);
                    document.removeEventListener('mouseup', handleMouseUp);
                    
                    // Calculate new grid position
                    const colWidth = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-col-width'));
                    const rowHeight = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-row-height'));
                    const marginX = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-margin-x'));
                    const marginY = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-margin-y'));
                    
                    const newX = Math.round((el.offsetLeft - marginX) / (colWidth + marginX));
                    const newY = Math.round((el.offsetTop - marginY) / (rowHeight + marginY));
                    
                    this.$emit('move', { i: this.i, x: newX, y: newY });
                    this.$parent.$emit('layout-updated');
                };
                
                document.addEventListener('mousemove', handleMouseMove);
                document.addEventListener('mouseup', handleMouseUp);
            });
        },
        
        startResize(e) {
            e.preventDefault();
            e.stopPropagation();
            
            const el = this.$el;
            const startWidth = el.offsetWidth;
            const startHeight = el.offsetHeight;
            const startX = e.clientX;
            const startY = e.clientY;
            
            el.classList.add('resizing');
            
            const handleMouseMove = (e) => {
                const dx = e.clientX - startX;
                const dy = e.clientY - startY;
                
                el.style.width = (startWidth + dx) + 'px';
                el.style.height = (startHeight + dy) + 'px';
            };
            
            const handleMouseUp = () => {
                el.classList.remove('resizing');
                document.removeEventListener('mousemove', handleMouseMove);
                document.removeEventListener('mouseup', handleMouseUp);
                
                // Calculate new grid size
                const colWidth = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-col-width'));
                const rowHeight = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-row-height'));
                const marginX = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-margin-x'));
                const marginY = parseFloat(getComputedStyle(el.parentElement).getPropertyValue('--grid-margin-y'));
                
                const newW = Math.round((el.offsetWidth + marginX) / (colWidth + marginX));
                const newH = Math.round((el.offsetHeight + marginY) / (rowHeight + marginY));
                
                this.$emit('resize', { i: this.i, w: newW, h: newH });
                this.$parent.$emit('layout-updated');
            };
            
            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);
        }
    }
};

// Export globally
if (typeof window !== 'undefined') {
    window.SimpleGridLayout = SimpleGridLayout;
    window.SimpleGridItem = SimpleGridItem;
}