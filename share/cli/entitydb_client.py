#!/usr/bin/env python3
"""
EntityDB API Client

A Python client for interacting with the EntityDB API
"""

import argparse
import json
import os
import sys
import requests
from typing import Dict, List, Optional, Union, Any


class EntityDBClient:
    """Client for interacting with the EntityDB API"""

    def __init__(self, server: str = "http://localhost:8085", token_file: str = None) -> None:
        """Initialize the client

        Args:
            server: The server URL
            token_file: Path to the token file
        """
        self.server = server
        if token_file is None:
            self.token_file = os.path.expanduser("~/.entitydb_token")
        else:
            self.token_file = os.path.expanduser(token_file)
        self.token = self._load_token()

    def _load_token(self) -> Optional[str]:
        """Load the authentication token

        Returns:
            The token or None if not found
        """
        if not os.path.exists(self.token_file):
            return None

        with open(self.token_file, "r") as f:
            return f.read().strip()

    def _save_token(self, token: str) -> None:
        """Save the authentication token

        Args:
            token: The token to save
        """
        with open(self.token_file, "w") as f:
            f.write(token)

    def _get_headers(self) -> Dict[str, str]:
        """Get the headers for API requests

        Returns:
            Dict of headers
        """
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers

    def _request(
        self, method: str, endpoint: str, data: Optional[Dict] = None
    ) -> Dict:
        """Make an API request

        Args:
            method: HTTP method
            endpoint: API endpoint
            data: Request data

        Returns:
            API response
        """
        url = f"{self.server}{endpoint}"
        headers = self._get_headers()

        response = requests.request(method, url, headers=headers, json=data)
        try:
            return response.json()
        except json.JSONDecodeError:
            return {"status": "error", "message": "Invalid JSON response"}

    def login(self, username: str, password: str) -> Dict:
        """Authenticate and get a token

        Args:
            username: Username
            password: Password

        Returns:
            Authentication response
        """
        data = {"username": username, "password": password}
        response = self._request("POST", "/api/v1/auth/login", data)

        # Extract token from the response
        token = None
        if response.get("status") == "ok":
            if "token" in response:
                token = response["token"]
            elif "data" in response and "token" in response["data"]:
                token = response["data"]["token"]

        if token:
            self._save_token(token)
            self.token = token
            return {"status": "ok", "message": "Authentication successful"}

        return response

    def list_entities(
        self,
        entity_type: Optional[str] = None,
        tags: Optional[List[str]] = None,
        limit: int = 10,
        offset: int = 0,
    ) -> Dict:
        """List entities

        Args:
            entity_type: Filter by entity type
            tags: Filter by tags
            limit: Result limit
            offset: Result offset

        Returns:
            List of entities
        """
        params = []
        if entity_type:
            params.append(f"type={entity_type}")
        if tags:
            params.append(f"tags={','.join(tags)}")
        if limit:
            params.append(f"limit={limit}")
        if offset:
            params.append(f"offset={offset}")

        endpoint = "/api/v1/entities"
        if params:
            endpoint = f"{endpoint}?{'&'.join(params)}"

        return self._request("GET", endpoint)

    def get_entity(self, entity_id: str) -> Dict:
        """Get entity details

        Args:
            entity_id: Entity ID

        Returns:
            Entity details
        """
        return self._request("GET", f"/api/v1/entities/{entity_id}")

    def create_entity(
        self,
        entity_type: str,
        title: str,
        description: Optional[str] = None,
        tags: Optional[List[str]] = None,
        properties: Optional[Dict] = None,
    ) -> Dict:
        """Create a new entity

        Args:
            entity_type: Entity type
            title: Entity title
            description: Entity description
            tags: Entity tags
            properties: Entity properties

        Returns:
            Creation response
        """
        data = {"type": entity_type, "title": title}
        if description:
            data["description"] = description
        if tags:
            data["tags"] = tags
        if properties:
            data["properties"] = properties

        return self._request("POST", "/api/v1/entities", data)

    def update_entity(
        self,
        entity_id: str,
        title: Optional[str] = None,
        description: Optional[str] = None,
        tags: Optional[List[str]] = None,
        properties: Optional[Dict] = None,
    ) -> Dict:
        """Update an entity

        Args:
            entity_id: Entity ID
            title: New entity title
            description: New entity description
            tags: New entity tags
            properties: New entity properties

        Returns:
            Update response
        """
        data = {}
        if title:
            data["title"] = title
        if description:
            data["description"] = description
        if tags:
            data["tags"] = tags
        if properties:
            data["properties"] = properties

        return self._request("PUT", f"/api/v1/entities/{entity_id}", data)

    def delete_entity(self, entity_id: str) -> Dict:
        """Delete an entity

        Args:
            entity_id: Entity ID

        Returns:
            Deletion response
        """
        return self._request("DELETE", f"/api/v1/entities/{entity_id}")

    def list_relationships(
        self,
        source_id: Optional[str] = None,
        target_id: Optional[str] = None,
        relationship_type: Optional[str] = None,
    ) -> Dict:
        """List relationships

        Args:
            source_id: Filter by source entity ID
            target_id: Filter by target entity ID
            relationship_type: Filter by relationship type

        Returns:
            List of relationships
        """
        params = []
        if source_id:
            params.append(f"source_id={source_id}")
        if target_id:
            params.append(f"target_id={target_id}")
        if relationship_type:
            params.append(f"relationship_type={relationship_type}")

        endpoint = "/api/v1/entity-relationships"
        if params:
            endpoint = f"{endpoint}?{'&'.join(params)}"

        return self._request("GET", endpoint)

    def create_relationship(
        self,
        source_id: str,
        target_id: str,
        relationship_type: str,
        properties: Optional[Dict] = None,
    ) -> Dict:
        """Create a relationship

        Args:
            source_id: Source entity ID
            target_id: Target entity ID
            relationship_type: Relationship type
            properties: Relationship properties

        Returns:
            Creation response
        """
        data = {
            "source_id": source_id,
            "target_id": target_id,
            "type": relationship_type,
        }
        if properties:
            data["properties"] = properties

        return self._request("POST", "/api/v1/entity-relationships", data)

    def delete_relationship(
        self, source_id: str, target_id: str, relationship_type: Optional[str] = None
    ) -> Dict:
        """Delete a relationship

        Args:
            source_id: Source entity ID
            target_id: Target entity ID
            relationship_type: Relationship type

        Returns:
            Deletion response
        """
        params = [f"source_id={source_id}", f"target_id={target_id}"]
        if relationship_type:
            params.append(f"relationship_type={relationship_type}")

        endpoint = f"/api/v1/entity-relationships?{'&'.join(params)}"
        return self._request("DELETE", endpoint)


def format_output(data: Dict, output_format: str = "table") -> str:
    """Format the output

    Args:
        data: Data to format
        output_format: Output format (table or json)

    Returns:
        Formatted output
    """
    if output_format == "json":
        return json.dumps(data, indent=2)

    # Format as table
    result = []
    if "status" in data:
        result.append(f"Status: {data['status']}")
    if "message" in data:
        result.append(f"Message: {data['message']}")

    # Format entities
    if "data" in data and isinstance(data["data"], list):
        result.append("\nEntities:")
        for entity in data["data"]:
            entity_id = entity.get("id", "unknown")
            title = entity.get("title", "Untitled")
            entity_type = entity.get("type", "unknown")
            result.append(f"  {entity_id} - {title} ({entity_type})")

            # Show tags if present
            if "tags" in entity and entity["tags"]:
                tags_str = ", ".join(entity["tags"])
                result.append(f"    Tags: {tags_str}")

    # Format relationships
    if "relationships" in data and isinstance(data["relationships"], list):
        result.append("\nRelationships:")
        for rel in data["relationships"]:
            source = rel.get("source_id", "unknown")
            target = rel.get("target_id", "unknown")
            rel_type = rel.get("type", "unknown")
            result.append(f"  {source} --[{rel_type}]--> {target}")

    return "\n".join(result)


def parse_tags(tags_str: Optional[str]) -> Optional[List[str]]:
    """Parse tags string into a list

    Args:
        tags_str: Comma-separated tags

    Returns:
        List of tags or None
    """
    if not tags_str:
        return None
    return [tag.strip() for tag in tags_str.split(",")]


def parse_properties(properties_str: Optional[str]) -> Optional[Dict]:
    """Parse properties string into a dict

    Args:
        properties_str: JSON properties string

    Returns:
        Properties dict or None
    """
    if not properties_str:
        return None
    try:
        return json.loads(properties_str)
    except json.JSONDecodeError:
        print("Error: Invalid JSON for properties", file=sys.stderr)
        sys.exit(1)


def main() -> None:
    """Main function"""
    parser = argparse.ArgumentParser(description="EntityDB API Client")
    parser.add_argument(
        "--server", default="http://localhost:8085", help="API server URL"
    )
    parser.add_argument(
        "--token-file", default=None, help="Token file location"
    )
    parser.add_argument(
        "--format",
        choices=["table", "json"],
        default="table",
        help="Output format",
    )

    subparsers = parser.add_subparsers(dest="command", help="Commands")

    # Login command
    login_parser = subparsers.add_parser("login", help="Authenticate and get a token")
    login_parser.add_argument("username", help="Username")
    login_parser.add_argument("password", help="Password")

    # Entity commands
    entity_parser = subparsers.add_parser("entity", help="Entity operations")
    entity_subparsers = entity_parser.add_subparsers(dest="subcommand")

    # Entity list
    entity_list_parser = entity_subparsers.add_parser("list", help="List entities")
    entity_list_parser.add_argument("--type", help="Filter by entity type")
    entity_list_parser.add_argument("--tags", help="Filter by tags (comma-separated)")
    entity_list_parser.add_argument(
        "--limit", type=int, default=10, help="Result limit"
    )
    entity_list_parser.add_argument(
        "--offset", type=int, default=0, help="Result offset"
    )

    # Entity get
    entity_get_parser = entity_subparsers.add_parser("get", help="Get entity details")
    entity_get_parser.add_argument("--id", required=True, help="Entity ID")

    # Entity create
    entity_create_parser = entity_subparsers.add_parser(
        "create", help="Create a new entity"
    )
    entity_create_parser.add_argument("--type", required=True, help="Entity type")
    entity_create_parser.add_argument("--title", required=True, help="Entity title")
    entity_create_parser.add_argument("--description", help="Entity description")
    entity_create_parser.add_argument("--tags", help="Entity tags (comma-separated)")
    entity_create_parser.add_argument(
        "--properties", help="Entity properties as JSON"
    )

    # Entity update
    entity_update_parser = entity_subparsers.add_parser("update", help="Update an entity")
    entity_update_parser.add_argument("--id", required=True, help="Entity ID")
    entity_update_parser.add_argument("--title", help="New entity title")
    entity_update_parser.add_argument("--description", help="New entity description")
    entity_update_parser.add_argument("--tags", help="New entity tags (comma-separated)")
    entity_update_parser.add_argument(
        "--properties", help="New entity properties as JSON"
    )

    # Entity delete
    entity_delete_parser = entity_subparsers.add_parser("delete", help="Delete an entity")
    entity_delete_parser.add_argument("--id", required=True, help="Entity ID")

    # Relationship commands
    relationship_parser = subparsers.add_parser("relationship", help="Relationship operations")
    relationship_subparsers = relationship_parser.add_subparsers(dest="subcommand")

    # Relationship list
    relationship_list_parser = relationship_subparsers.add_parser(
        "list", help="List relationships"
    )
    relationship_list_parser.add_argument("--source", help="Filter by source entity ID")
    relationship_list_parser.add_argument("--target", help="Filter by target entity ID")
    relationship_list_parser.add_argument("--type", help="Filter by relationship type")

    # Relationship create
    relationship_create_parser = relationship_subparsers.add_parser(
        "create", help="Create a relationship"
    )
    relationship_create_parser.add_argument(
        "--source", required=True, help="Source entity ID"
    )
    relationship_create_parser.add_argument(
        "--target", required=True, help="Target entity ID"
    )
    relationship_create_parser.add_argument(
        "--type", required=True, help="Relationship type"
    )
    relationship_create_parser.add_argument(
        "--properties", help="Relationship properties as JSON"
    )

    # Relationship delete
    relationship_delete_parser = relationship_subparsers.add_parser(
        "delete", help="Delete a relationship"
    )
    relationship_delete_parser.add_argument(
        "--source", required=True, help="Source entity ID"
    )
    relationship_delete_parser.add_argument(
        "--target", required=True, help="Target entity ID"
    )
    relationship_delete_parser.add_argument("--type", help="Relationship type")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    client = EntityDBClient(server=args.server, token_file=args.token_file)

    # Process commands
    result = None

    if args.command == "login":
        result = client.login(args.username, args.password)

    elif args.command == "entity":
        if args.subcommand == "list":
            result = client.list_entities(
                entity_type=args.type,
                tags=parse_tags(args.tags),
                limit=args.limit,
                offset=args.offset,
            )
        elif args.subcommand == "get":
            result = client.get_entity(args.id)
        elif args.subcommand == "create":
            result = client.create_entity(
                entity_type=args.type,
                title=args.title,
                description=args.description,
                tags=parse_tags(args.tags),
                properties=parse_properties(args.properties),
            )
        elif args.subcommand == "update":
            result = client.update_entity(
                entity_id=args.id,
                title=args.title,
                description=args.description,
                tags=parse_tags(args.tags),
                properties=parse_properties(args.properties),
            )
        elif args.subcommand == "delete":
            result = client.delete_entity(args.id)
        else:
            print(f"Unknown entity subcommand: {args.subcommand}", file=sys.stderr)
            sys.exit(1)

    elif args.command == "relationship":
        if args.subcommand == "list":
            result = client.list_relationships(
                source_id=args.source,
                target_id=args.target,
                relationship_type=args.type,
            )
        elif args.subcommand == "create":
            result = client.create_relationship(
                source_id=args.source,
                target_id=args.target,
                relationship_type=args.type,
                properties=parse_properties(args.properties),
            )
        elif args.subcommand == "delete":
            result = client.delete_relationship(
                source_id=args.source,
                target_id=args.target,
                relationship_type=args.type,
            )
        else:
            print(f"Unknown relationship subcommand: {args.subcommand}", file=sys.stderr)
            sys.exit(1)

    # Print result
    if result:
        print(format_output(result, args.format))


if __name__ == "__main__":
    main()