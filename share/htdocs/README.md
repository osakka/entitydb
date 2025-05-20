# EntityDB Web Interface

A clean, modern web interface for the EntityDB (EntityDB) platform built with Alpine.js.

## Overview

This web interface provides a simple, effective UI for interacting with the EntityDB entity API. It uses Alpine.js for reactive functionality with minimal JavaScript complexity.

## Features

- **JWT Authentication**: Secure login system with token-based authentication
- **Entity Management**: Create, list, and filter entities
- **Responsive Design**: Mobile-friendly interface that works on all devices
- **Real-time Filtering**: Search and filter entities by type and keywords
- **Clean Architecture**: Minimal dependencies, Alpine.js for reactivity

## Technology Stack

- **Alpine.js**: Lightweight reactive framework for modern UIs
- **Pure CSS**: Custom styling without external frameworks
- **JWT**: Secure authentication with localStorage persistence
- **RESTful API**: Clean integration with the entity API

## File Structure

```
/opt/entitydb/share/htdocs/
├── index.html       # Main HTML file with Alpine.js components
├── css/
│   └── style.css    # Custom CSS styling
├── js/
│   └── app.js       # Alpine.js component logic
└── README.md        # This file
```

## Usage

1. Access the interface at `http://localhost:8085/`
2. Login with your credentials (default: admin/password)
3. Use the interface to:
   - View all entities
   - Filter by type or search term
   - Create new entities
   - View entity details

## API Integration

The interface connects to these main API endpoints:

- `POST /api/v1/auth/login` - User authentication
- `GET /api/v1/entities/list` - List entities with optional filters
- `POST /api/v1/entities/create` - Create new entities

All requests use JWT authentication with the token stored in localStorage.

## Development

### Running Locally

The interface is served directly by the EntityDB server:

```bash
# Start the server
/opt/entitydb/bin/entitydbd.sh start

# Access the interface
open http://localhost:8085/
```

### Making Changes

1. Edit files in `/opt/entitydb/share/htdocs/`
2. Refresh your browser to see changes (no build process required)
3. Test across different browsers and devices

### Adding Features

To add new features:

1. Update the Alpine.js component in `js/app.js`
2. Add corresponding HTML with Alpine.js directives in `index.html`
3. Style with CSS in `css/style.css`

## Architecture Decisions

1. **Alpine.js Only**: We chose Alpine.js for its simplicity and minimal overhead
2. **No Build Process**: Direct serving of files for easy development
3. **Pure CSS**: Custom styling without CSS frameworks for full control
4. **JWT in localStorage**: Simple but effective authentication persistence
5. **RESTful API**: Clean separation of concerns with the backend

## Future Enhancements

Potential improvements for the interface:

- [ ] Add HTMX for server-side rendering capabilities
- [ ] Implement entity editing and deletion
- [ ] Add entity relationship visualization
- [ ] Include real-time updates via WebSocket
- [ ] Add dark mode toggle
- [ ] Implement entity search pagination

## Troubleshooting

### Common Issues

1. **Login fails**: Check server is running and credentials are correct
2. **Entities don't load**: Verify API endpoints are accessible
3. **Styling broken**: Clear browser cache or check CSS path
4. **JavaScript errors**: Check browser console for Alpine.js issues

### Debug Mode

Enable debug mode by setting `localStorage.debug = true` in browser console.

## License

This interface is part of the EntityDB project and follows the same licensing terms.