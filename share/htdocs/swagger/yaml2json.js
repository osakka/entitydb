// A simple YAML to JSON converter
const fs = require('fs');
const yaml = require('js-yaml');

try {
  const yamlContent = fs.readFileSync('swagger.yaml', 'utf8');
  const jsonContent = yaml.load(yamlContent);
  fs.writeFileSync('swagger.json', JSON.stringify(jsonContent, null, 2));
  console.log('Successfully converted swagger.yaml to swagger.json');
} catch (e) {
  console.error('Error:', e.message);
}