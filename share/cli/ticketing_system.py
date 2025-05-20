#!/usr/bin/env python3
"""
EntityDB Ticketing System Example

This demonstrates how to build a ticketing system using EntityDB's temporal database.
"""

import requests
import json
from datetime import datetime

class TicketingSystem:
    def __init__(self, base_url="http://localhost:8085/api/v1"):
        self.base_url = base_url
        self.token = None
        
    def login(self, username, password):
        """Login and get authentication token"""
        response = requests.post(f"{self.base_url}/auth/login",
                                json={"username": username, "password": password})
        response.raise_for_status()
        self.token = response.json()['token']
        print(f"✓ Logged in as {username}")
        
    def _headers(self):
        """Get headers with authentication"""
        return {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json"
        }
        
    def create_project(self, code, name, description):
        """Create a project entity"""
        entity = {
            "tags": [
                "type:project",
                f"id:code:{code}",
                f"name:{name}",
                "status:active"
            ],
            "content": [
                {"type": "description", "value": description}
            ]
        }
        
        response = requests.post(f"{self.base_url}/entities/create",
                                headers=self._headers(),
                                json=entity)
        response.raise_for_status()
        project = response.json()
        print(f"✓ Created project: {code} - {name}")
        return project['id']
        
    def create_ticket(self, project_code, ticket_id, title, description, 
                     priority="medium", status="open", assigned_to=None):
        """Create a ticket entity"""
        tags = [
            "type:ticket",
            f"id:ticket:{ticket_id}",
            f"project:{project_code}",
            f"status:{status}",
            f"priority:{priority}",
            "created_by:admin"
        ]
        
        if assigned_to:
            tags.append(f"assigned_to:{assigned_to}")
            
        entity = {
            "tags": tags,
            "content": [
                {"type": "title", "value": title},
                {"type": "description", "value": description},
                {"type": "created_at", "value": datetime.now().isoformat()}
            ]
        }
        
        response = requests.post(f"{self.base_url}/entities/create",
                                headers=self._headers(),
                                json=entity)
        response.raise_for_status()
        ticket = response.json()
        print(f"✓ Created ticket: {ticket_id} - {title}")
        return ticket['id']
        
    def add_comment(self, ticket_id, comment_text, author="admin"):
        """Add a comment to a ticket"""
        entity = {
            "tags": [
                "type:comment",
                f"ticket:{ticket_id}",
                f"author:{author}"
            ],
            "content": [
                {"type": "text", "value": comment_text},
                {"type": "timestamp", "value": datetime.now().isoformat()}
            ]
        }
        
        response = requests.post(f"{self.base_url}/entities/create",
                                headers=self._headers(),
                                json=entity)
        response.raise_for_status()
        print(f"✓ Added comment to {ticket_id}")
        
    def update_ticket_status(self, entity_id, new_status):
        """Update ticket status"""
        # First get the current entity
        response = requests.get(f"{self.base_url}/entities/get?id={entity_id}",
                               headers=self._headers())
        response.raise_for_status()
        ticket = response.json()
        
        # Update the status tag
        new_tags = []
        for tag in ticket['tags']:
            if tag.startswith('status:'):
                new_tags.append(f'status:{new_status}')
            else:
                new_tags.append(tag)
                
        # Update the entity
        update_data = {
            "id": entity_id,
            "tags": new_tags,
            "content": ticket['content']
        }
        
        response = requests.put(f"{self.base_url}/entities/update",
                               headers=self._headers(),
                               json=update_data)
        response.raise_for_status()
        print(f"✓ Updated ticket {entity_id} status to {new_status}")
        
    def list_tickets(self, project_code=None, status=None):
        """List tickets with optional filtering"""
        params = {}
        if project_code:
            params['tag'] = f'project:{project_code}'
        elif status:
            params['tag'] = f'status:{status}'
        else:
            params['tag'] = 'type:ticket'
            
        response = requests.get(f"{self.base_url}/entities/list",
                               headers=self._headers(),
                               params=params)
        response.raise_for_status()
        tickets = response.json()
        
        print(f"\\n=== Tickets ({len(tickets)} found) ===")
        for ticket in tickets:
            # Extract ticket info from tags
            ticket_id = None
            status = None
            priority = None
            
            for tag in ticket['tags']:
                if tag.startswith('id:ticket:'):
                    ticket_id = tag.split(':')[2]
                elif tag.startswith('status:'):
                    status = tag.split(':')[1]
                elif tag.startswith('priority:'):
                    priority = tag.split(':')[1]
                    
            # Get title from content
            title = "No title"
            for content in ticket['content']:
                if content['type'] == 'title':
                    title = content['value']
                    break
                    
            print(f"• {ticket_id}: {title} [{status}] ({priority})")
        
        return tickets

def demo():
    """Run a demo of the ticketing system"""
    ts = TicketingSystem()
    
    # Login
    ts.login("admin", "admin")
    
    # Create a project
    project_id = ts.create_project("WEBAPP", "Web Application", 
                                  "Main web application project")
    
    # Create some tickets
    ticket1_id = ts.create_ticket("WEBAPP", "WEBAPP-001", 
                                 "Login page not responsive on mobile",
                                 "Users report that the login page doesn't work properly on mobile devices",
                                 priority="high")
    
    ticket2_id = ts.create_ticket("WEBAPP", "WEBAPP-002",
                                 "Add dark mode support",
                                 "Many users have requested a dark mode option for the application",
                                 priority="medium")
    
    ticket3_id = ts.create_ticket("WEBAPP", "WEBAPP-003",
                                 "Database connection timeout",
                                 "Connection to database times out during peak hours",
                                 priority="critical",
                                 status="in_progress")
    
    # Add comments
    ts.add_comment("WEBAPP-001", "Confirmed the issue on iPhone 12 and Samsung Galaxy S21")
    ts.add_comment("WEBAPP-001", "Working on responsive CSS fixes")
    ts.add_comment("WEBAPP-003", "Increased connection pool size, monitoring results")
    
    # Update ticket status
    ts.update_ticket_status(ticket1_id, "in_progress")
    
    # List all tickets
    ts.list_tickets()
    
    # List tickets by status
    print("\\n=== Open Tickets ===")
    ts.list_tickets(status="open")
    
    print("\\n=== In Progress Tickets ===")
    ts.list_tickets(status="in_progress")

if __name__ == "__main__":
    demo()