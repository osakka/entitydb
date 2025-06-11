#!/bin/bash
# Systematic rename of hub to dataset

echo "=== Renaming Hub to Dataset ==="
echo

# Step 1: Rename Go files
echo "Step 1: Renaming Go files..."
cd /opt/entitydb/src/api

# Rename files
mv hub_management_handler.go dataset_management_handler.go 2>/dev/null && echo "✓ Renamed hub_management_handler.go"
mv hub_management_handler_rbac.go dataset_management_handler_rbac.go 2>/dev/null && echo "✓ Renamed hub_management_handler_rbac.go"
mv hub_entity_handler_rbac.go dataset_entity_handler_rbac.go 2>/dev/null && echo "✓ Renamed hub_entity_handler_rbac.go"
# entity_handler_hub.go already became entity_handler_dataset.go
rm -f entity_handler_hub.go 2>/dev/null
mv hub_middleware.go dataset_middleware.go 2>/dev/null && echo "✓ Renamed hub_middleware.go"

cd /opt/entitydb

# Step 2: Update Go source code
echo
echo "Step 2: Updating Go source code..."

# Function to update a file
update_file() {
    local file=$1
    if [ -f "$file" ]; then
        # Type and struct names
        sed -i 's/HubManagementHandler/DatasetManagementHandler/g' "$file"
        sed -i 's/hubManagementHandler/datasetManagementHandler/g' "$file"
        sed -i 's/CreateHubRequest/CreateDatasetRequest/g' "$file"
        sed -i 's/CreateHubResponse/CreateDatasetResponse/g' "$file"
        sed -i 's/HubInfo/DatasetInfo/g' "$file"
        sed -i 's/HubContext/DatasetContext/g' "$file"
        sed -i 's/HubMiddleware/DatasetMiddleware/g' "$file"
        
        # Function names
        sed -i 's/CreateHub/CreateDataset/g' "$file"
        sed -i 's/DeleteHub/DeleteDataset/g' "$file"
        sed -i 's/ListHubs/ListDatasets/g' "$file"
        sed -i 's/GetHub/GetDataset/g' "$file"
        sed -i 's/FormatHubTag/FormatDatasetTag/g' "$file"
        sed -i 's/ParseHubTag/ParseDatasetTag/g' "$file"
        sed -i 's/CheckHubPermission/CheckDatasetPermission/g' "$file"
        sed -i 's/ValidateEntityHub/ValidateEntityDataset/g' "$file"
        sed -i 's/GetHubContext/GetDatasetContext/g' "$file"
        sed -i 's/QueryHubEntities/QueryDatasetEntities/g' "$file"
        sed -i 's/CreateHubEntity/CreateDatasetEntity/g' "$file"
        
        # Variables and parameters
        sed -i 's/hubName/datasetName/g' "$file"
        sed -i 's/hub_name/dataset_name/g' "$file"
        sed -i 's/HubName/DatasetName/g' "$file"
        sed -i 's/currentHub/currentDataset/g' "$file"
        sed -i 's/targetHub/targetDataset/g' "$file"
        
        # String literals and tags
        sed -i 's/"hub:/"dataset:/g' "$file"
        sed -i 's/`hub:/`dataset:/g' "$file"
        sed -i 's/"type:hub"/"type:dataset"/g' "$file"
        sed -i 's/"hub_name:/"dataset_name:/g' "$file"
        sed -i 's/hub\/entities/dataset\/entities/g' "$file"
        
        # API paths
        sed -i 's|/hubs/|/datasets/|g' "$file"
        sed -i 's|/hub/|/dataset/|g' "$file"
        
        # RBAC permissions
        sed -i 's/:hub:/:dataset:/g' "$file"
        sed -i 's/rbac:perm:hub:/rbac:perm:dataset:/g' "$file"
        
        # Error messages and logs
        sed -i 's/Hub not found/Dataset not found/g' "$file"
        sed -i 's/hub not found/dataset not found/g' "$file"
        sed -i 's/Failed to create hub/Failed to create dataset/g' "$file"
        sed -i 's/Invalid hub/Invalid dataset/g' "$file"
        sed -i 's/hub entities/dataset entities/g' "$file"
        sed -i 's/Hub entities/Dataset entities/g' "$file"
        
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
        sed -i 's/hub:/dataset:/g' "$file"
        sed -i 's/Hub/Dataset/g' "$file"
        sed -i 's/hub/dataset/g' "$file"
        sed -i 's/HUB/DATASPACE/g' "$file"
    fi
done

# Update README
sed -i 's|/hubs/|/datasets/|g' README.md
sed -i 's/hub:/dataset:/g' README.md

# Step 4: Update frontend files
echo
echo "Step 4: Updating frontend files..."

# Update Worca application
if [ -f "share/htdocs/worca/worca.js" ]; then
    sed -i 's/transformHubEntities/transformDatasetEntities/g' share/htdocs/worca/worca.js
    sed -i 's/hub:/dataset:/g' share/htdocs/worca/worca.js
    echo "✓ Updated worca.js"
fi

if [ -f "share/htdocs/worca/worca-api.js" ]; then
    sed -i 's/hub:/dataset:/g' share/htdocs/worca/worca-api.js
    sed -i 's/hub format/dataset format/g' share/htdocs/worca/worca-api.js
    echo "✓ Updated worca-api.js"
fi

echo
echo "=== Rename Complete ==="
echo
echo "Next steps:"
echo "1. Run 'make' to ensure compilation"
echo "2. Run tests to verify functionality"
echo "3. Update any remaining references manually"