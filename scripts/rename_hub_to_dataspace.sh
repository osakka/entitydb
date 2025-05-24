#!/bin/bash
# Systematic rename of hub to dataspace

echo "=== Renaming Hub to Dataspace ==="
echo

# Step 1: Rename Go files
echo "Step 1: Renaming Go files..."
cd /opt/entitydb/src/api

# Rename files
mv hub_management_handler.go dataspace_management_handler.go 2>/dev/null && echo "✓ Renamed hub_management_handler.go"
mv hub_management_handler_rbac.go dataspace_management_handler_rbac.go 2>/dev/null && echo "✓ Renamed hub_management_handler_rbac.go"
mv hub_entity_handler_rbac.go dataspace_entity_handler_rbac.go 2>/dev/null && echo "✓ Renamed hub_entity_handler_rbac.go"
# entity_handler_hub.go already became entity_handler_dataspace.go
rm -f entity_handler_hub.go 2>/dev/null
mv hub_middleware.go dataspace_middleware.go 2>/dev/null && echo "✓ Renamed hub_middleware.go"

cd /opt/entitydb

# Step 2: Update Go source code
echo
echo "Step 2: Updating Go source code..."

# Function to update a file
update_file() {
    local file=$1
    if [ -f "$file" ]; then
        # Type and struct names
        sed -i 's/HubManagementHandler/DataspaceManagementHandler/g' "$file"
        sed -i 's/hubManagementHandler/dataspaceManagementHandler/g' "$file"
        sed -i 's/CreateHubRequest/CreateDataspaceRequest/g' "$file"
        sed -i 's/CreateHubResponse/CreateDataspaceResponse/g' "$file"
        sed -i 's/HubInfo/DataspaceInfo/g' "$file"
        sed -i 's/HubContext/DataspaceContext/g' "$file"
        sed -i 's/HubMiddleware/DataspaceMiddleware/g' "$file"
        
        # Function names
        sed -i 's/CreateHub/CreateDataspace/g' "$file"
        sed -i 's/DeleteHub/DeleteDataspace/g' "$file"
        sed -i 's/ListHubs/ListDataspaces/g' "$file"
        sed -i 's/GetHub/GetDataspace/g' "$file"
        sed -i 's/FormatHubTag/FormatDataspaceTag/g' "$file"
        sed -i 's/ParseHubTag/ParseDataspaceTag/g' "$file"
        sed -i 's/CheckHubPermission/CheckDataspacePermission/g' "$file"
        sed -i 's/ValidateEntityHub/ValidateEntityDataspace/g' "$file"
        sed -i 's/GetHubContext/GetDataspaceContext/g' "$file"
        sed -i 's/QueryHubEntities/QueryDataspaceEntities/g' "$file"
        sed -i 's/CreateHubEntity/CreateDataspaceEntity/g' "$file"
        
        # Variables and parameters
        sed -i 's/hubName/dataspaceName/g' "$file"
        sed -i 's/hub_name/dataspace_name/g' "$file"
        sed -i 's/HubName/DataspaceName/g' "$file"
        sed -i 's/currentHub/currentDataspace/g' "$file"
        sed -i 's/targetHub/targetDataspace/g' "$file"
        
        # String literals and tags
        sed -i 's/"hub:/"dataspace:/g' "$file"
        sed -i 's/`hub:/`dataspace:/g' "$file"
        sed -i 's/"type:hub"/"type:dataspace"/g' "$file"
        sed -i 's/"hub_name:/"dataspace_name:/g' "$file"
        sed -i 's/hub\/entities/dataspace\/entities/g' "$file"
        
        # API paths
        sed -i 's|/hubs/|/dataspaces/|g' "$file"
        sed -i 's|/hub/|/dataspace/|g' "$file"
        
        # RBAC permissions
        sed -i 's/:hub:/:dataspace:/g' "$file"
        sed -i 's/rbac:perm:hub:/rbac:perm:dataspace:/g' "$file"
        
        # Error messages and logs
        sed -i 's/Hub not found/Dataspace not found/g' "$file"
        sed -i 's/hub not found/dataspace not found/g' "$file"
        sed -i 's/Failed to create hub/Failed to create dataspace/g' "$file"
        sed -i 's/Invalid hub/Invalid dataspace/g' "$file"
        sed -i 's/hub entities/dataspace entities/g' "$file"
        sed -i 's/Hub entities/Dataspace entities/g' "$file"
        
        echo "✓ Updated $file"
    fi
}

# Update all Go files
for file in $(find src -name "*.go" -type f); do
    update_file "$file"
done

# Step 3: Update documentation
echo
echo "Step 3: Updating documentation..."

# Rename documentation files
mv docs/implementation/MULTI_HUB_ARCHITECTURE.md docs/implementation/MULTI_DATASPACE_ARCHITECTURE.md 2>/dev/null && echo "✓ Renamed MULTI_HUB_ARCHITECTURE.md"
mv docs/applications/worca/WORCHA_HUB_ARCHITECTURE.md docs/applications/worca/WORCA_DATASPACE_ARCHITECTURE.md 2>/dev/null && echo "✓ Renamed WORCHA_HUB_ARCHITECTURE.md"

# Update documentation content
for file in $(find docs -name "*.md" -type f); do
    if [ -f "$file" ]; then
        sed -i 's/hub:/dataspace:/g' "$file"
        sed -i 's/Hub/Dataspace/g' "$file"
        sed -i 's/hub/dataspace/g' "$file"
        sed -i 's/HUB/DATASPACE/g' "$file"
    fi
done

# Update README
sed -i 's|/hubs/|/dataspaces/|g' README.md
sed -i 's/hub:/dataspace:/g' README.md

# Step 4: Update frontend files
echo
echo "Step 4: Updating frontend files..."

# Update Worca application
if [ -f "share/htdocs/worca/worca.js" ]; then
    sed -i 's/transformHubEntities/transformDataspaceEntities/g' share/htdocs/worca/worca.js
    sed -i 's/hub:/dataspace:/g' share/htdocs/worca/worca.js
    echo "✓ Updated worca.js"
fi

if [ -f "share/htdocs/worca/worca-api.js" ]; then
    sed -i 's/hub:/dataspace:/g' share/htdocs/worca/worca-api.js
    sed -i 's/hub format/dataspace format/g' share/htdocs/worca/worca-api.js
    echo "✓ Updated worca-api.js"
fi

echo
echo "=== Rename Complete ==="
echo
echo "Next steps:"
echo "1. Run 'make' to ensure compilation"
echo "2. Run tests to verify functionality"
echo "3. Update any remaining references manually"