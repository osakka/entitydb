#!/usr/bin/env python3
"""
EntityDB Monitoring System Example
==================================

Demonstrates EntityDB's temporal database capabilities with a comprehensive
monitoring system that tracks servers, services, and metrics over time.

Features:
- Real-time metric collection and storage
- Historical trend analysis using temporal queries
- Intelligent alerting based on historical patterns
- Server and service health monitoring
- Performance baselines and anomaly detection

This showcases EntityDB's nanosecond-precision temporal capabilities!
"""

import requests
import json
import time
import random
import threading
from datetime import datetime, timedelta
from typing import Dict, List, Optional
from dataclasses import dataclass
import base64

class EntityDBClient:
    """EntityDB client for monitoring system"""
    
    def __init__(self, base_url: str = "https://localhost:8085", username: str = "admin", password: str = "admin"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.verify = False  # Skip SSL verification for demo
        self.token = None
        self.authenticate(username, password)
    
    def authenticate(self, username: str, password: str):
        """Authenticate and get token"""
        response = self.session.post(
            f"{self.base_url}/api/v1/auth/login",
            json={"username": username, "password": password}
        )
        if response.status_code == 200:
            self.token = response.json()["token"]
            self.session.headers.update({"Authorization": f"Bearer {self.token}"})
            print(f"✓ Authenticated with EntityDB")
        else:
            raise Exception(f"Authentication failed: {response.text}")
    
    def refresh_auth_if_needed(self):
        """Re-authenticate if needed"""
        # Simple test request to check if auth is still valid
        response = self.session.get(f"{self.base_url}/health")
        if response.status_code == 401:
            print("🔄 Re-authenticating...")
            self.authenticate("admin", "admin")
    
    def create_entity(self, entity_type: str, dataset: str, tags: Dict[str, str], content: str = "") -> str:
        """Create an entity with tags"""
        self.refresh_auth_if_needed()
        tag_list = [f"{k}:{v}" for k, v in tags.items()]
        tag_list.extend([f"type:{entity_type}", f"dataset:{dataset}"])
        
        content_b64 = base64.b64encode(content.encode()).decode() if content else ""
        
        response = self.session.post(
            f"{self.base_url}/api/v1/entities/create",
            json={
                "tags": tag_list,
                "content": content_b64
            }
        )
        if response.status_code == 201:
            return response.json()["id"]
        else:
            raise Exception(f"Failed to create entity: {response.status_code} - {response.text}")
    
    def add_metric_value(self, entity_id: str, metric_name: str, value: float, timestamp: Optional[str] = None):
        """Add a metric value as a temporal tag"""
        self.refresh_auth_if_needed()
        if timestamp is None:
            timestamp = datetime.now().isoformat()
        
        # Get current entity to add metric value
        entity_response = self.session.get(f"{self.base_url}/api/v1/entities/get?id={entity_id}")
        if entity_response.status_code != 200:
            print(f"Warning: Failed to get entity for metric update: {entity_response.text}")
            return
        
        entity = entity_response.json()
        current_tags = entity.get("tags", [])
        
        # Add value tag with timestamp for temporal tracking
        tag = f"value:{value}"
        updated_tags = current_tags + [tag]
        
        # Update entity with new metric value
        update_data = {
            "id": entity_id,
            "tags": updated_tags,
            "content": entity["content"]
        }
        
        response = self.session.put(f"{self.base_url}/api/v1/entities/update", json=update_data)
        if response.status_code != 200:
            print(f"Warning: Failed to add metric value: {response.text}")
    
    def query_entities(self, tag: str) -> List[Dict]:
        """Query entities by tag"""
        self.refresh_auth_if_needed()
        response = self.session.get(
            f"{self.base_url}/api/v1/entities/query",
            params={"tag": tag}
        )
        if response.status_code == 200:
            return response.json().get("entities", [])
        return []
    
    def get_entity_history(self, entity_id: str, hours_back: int = 24) -> List[Dict]:
        """Get entity history using temporal queries"""
        self.refresh_auth_if_needed()
        response = self.session.get(
            f"{self.base_url}/api/v1/entities/history",
            params={"id": entity_id}
        )
        if response.status_code == 200:
            return response.json()
        return []
    
    def get_entity_as_of(self, entity_id: str, timestamp: str) -> Optional[Dict]:
        """Get entity state as of specific timestamp"""
        self.refresh_auth_if_needed()
        response = self.session.get(
            f"{self.base_url}/api/v1/entities/as-of",
            params={"id": entity_id, "timestamp": timestamp}
        )
        if response.status_code == 200:
            return response.json()
        return None

@dataclass
class ServerMetrics:
    """Server monitoring metrics"""
    cpu_usage: float
    memory_usage: float
    disk_usage: float
    network_in: float
    network_out: float
    load_average: float
    active_connections: int

@dataclass
class ServiceMetrics:
    """Service monitoring metrics"""
    response_time: float
    requests_per_second: float
    error_rate: float
    availability: float
    active_users: int

class MonitoringSystem:
    """Comprehensive monitoring system using EntityDB temporal capabilities"""
    
    def __init__(self):
        self.client = EntityDBClient()
        self.servers: Dict[str, str] = {}  # hostname -> entity_id
        self.services: Dict[str, str] = {}  # service_name -> entity_id
        self.metrics: Dict[str, str] = {}  # metric_name -> entity_id
        self.running = False
        self.setup_infrastructure()
    
    def setup_infrastructure(self):
        """Setup monitoring infrastructure entities"""
        print("🏗️  Setting up monitoring infrastructure...")
        
        # Create server entities
        server_hosts = ["web-01", "web-02", "api-01", "api-02", "db-01", "cache-01"]
        for hostname in server_hosts:
            server_id = self.client.create_entity(
                entity_type="server",
                dataset="monitoring",
                tags={
                    "hostname": hostname,
                    "role": "web" if "web" in hostname else "api" if "api" in hostname else "database" if "db" in hostname else "cache",
                    "status": "active",
                    "environment": "production"
                },
                content=json.dumps({
                    "ip_address": f"10.0.1.{10 + len(self.servers)}",
                    "os": "Ubuntu 22.04",
                    "specs": {"cpu": "8 cores", "memory": "32GB", "disk": "500GB SSD"}
                })
            )
            self.servers[hostname] = server_id
            print(f"  ✓ Created server: {hostname}")
        
        # Create service entities
        services = ["web-frontend", "user-api", "product-api", "payment-api", "database", "redis-cache"]
        for service_name in services:
            service_id = self.client.create_entity(
                entity_type="service",
                dataset="monitoring", 
                tags={
                    "service_name": service_name,
                    "status": "healthy",
                    "version": "v2.1.0",
                    "critical": "true" if service_name in ["payment-api", "database"] else "false"
                },
                content=json.dumps({
                    "description": f"Production {service_name} service",
                    "health_check_url": f"https://{service_name}.company.com/health",
                    "dependencies": ["database"] if service_name != "database" else []
                })
            )
            self.services[service_name] = service_id
            print(f"  ✓ Created service: {service_name}")
        
        print(f"🎯 Infrastructure ready: {len(self.servers)} servers, {len(self.services)} services")
    
    def generate_server_metrics(self, hostname: str) -> ServerMetrics:
        """Generate realistic server metrics with trends and anomalies"""
        base_time = time.time() % 86400  # 24-hour cycle
        
        # CPU usage with daily patterns + random spikes
        cpu_base = 20 + 30 * abs(math.sin(base_time / 86400 * 2 * math.pi))  # Daily cycle
        cpu_spike = random.gauss(0, 5) if random.random() < 0.1 else 0  # 10% chance of spike
        cpu_usage = max(5, min(95, cpu_base + cpu_spike))
        
        # Memory usage with gradual increase (memory leaks simulation)
        memory_base = 40 + (base_time / 86400) * 20  # Gradual increase over day
        memory_usage = max(20, min(90, memory_base + random.gauss(0, 3)))
        
        # Disk usage (slowly increasing)
        disk_usage = 45 + random.gauss(0, 2)
        
        # Network traffic with business hours pattern
        business_hours = 9 <= (base_time / 3600) <= 17
        network_multiplier = 3.0 if business_hours else 0.5
        network_in = random.gauss(100, 20) * network_multiplier
        network_out = random.gauss(80, 15) * network_multiplier
        
        # Load average correlates with CPU
        load_average = cpu_usage / 100 * 8 + random.gauss(0, 0.5)
        
        # Active connections
        connections = int(random.gauss(150, 30) * network_multiplier)
        
        return ServerMetrics(
            cpu_usage=round(cpu_usage, 2),
            memory_usage=round(memory_usage, 2),
            disk_usage=round(disk_usage, 2),
            network_in=round(network_in, 2),
            network_out=round(network_out, 2),
            load_average=round(load_average, 2),
            active_connections=max(0, connections)
        )
    
    def generate_service_metrics(self, service_name: str) -> ServiceMetrics:
        """Generate realistic service metrics"""
        # Response time with occasional slowdowns
        base_response = 50 if "api" in service_name else 100 if service_name == "database" else 20
        response_spike = random.gauss(0, 20) if random.random() < 0.05 else 0
        response_time = max(10, base_response + response_spike)
        
        # RPS with business hours pattern
        base_time = time.time() % 86400
        business_hours = 9 <= (base_time / 3600) <= 17
        rps_multiplier = 2.0 if business_hours else 0.3
        requests_per_second = random.gauss(100, 20) * rps_multiplier
        
        # Error rate (usually low, occasionally spikes)
        error_rate = random.gauss(0.5, 0.2) if random.random() < 0.9 else random.gauss(5, 2)
        error_rate = max(0, min(20, error_rate))
        
        # Availability (high, with rare outages)
        availability = 99.9 if random.random() < 0.99 else random.uniform(85, 99)
        
        # Active users
        active_users = int(random.gauss(500, 100) * rps_multiplier)
        
        return ServiceMetrics(
            response_time=round(response_time, 2),
            requests_per_second=round(requests_per_second, 2),
            error_rate=round(error_rate, 3),
            availability=round(availability, 2),
            active_users=max(0, active_users)
        )
    
    def collect_metrics(self):
        """Collect and store metrics using EntityDB temporal capabilities"""
        print("📊 Starting metric collection...")
        
        while self.running:
            try:
                timestamp = datetime.now().isoformat()
                
                # Collect server metrics
                for hostname, server_id in self.servers.items():
                    metrics = self.generate_server_metrics(hostname)
                    
                    # Store each metric as a temporal tag
                    self.client.add_metric_value(server_id, "cpu_usage", metrics.cpu_usage, timestamp)
                    self.client.add_metric_value(server_id, "memory_usage", metrics.memory_usage, timestamp)
                    self.client.add_metric_value(server_id, "disk_usage", metrics.disk_usage, timestamp)
                    self.client.add_metric_value(server_id, "network_in", metrics.network_in, timestamp)
                    self.client.add_metric_value(server_id, "network_out", metrics.network_out, timestamp)
                    self.client.add_metric_value(server_id, "load_average", metrics.load_average, timestamp)
                    self.client.add_metric_value(server_id, "active_connections", metrics.active_connections, timestamp)
                
                # Collect service metrics
                for service_name, service_id in self.services.items():
                    metrics = self.generate_service_metrics(service_name)
                    
                    self.client.add_metric_value(service_id, "response_time", metrics.response_time, timestamp)
                    self.client.add_metric_value(service_id, "requests_per_second", metrics.requests_per_second, timestamp)
                    self.client.add_metric_value(service_id, "error_rate", metrics.error_rate, timestamp)
                    self.client.add_metric_value(service_id, "availability", metrics.availability, timestamp)
                    self.client.add_metric_value(service_id, "active_users", metrics.active_users, timestamp)
                
                print(f"📈 Collected metrics at {timestamp[:19]}")
                time.sleep(10)  # Collect every 10 seconds
                
            except Exception as e:
                print(f"❌ Error collecting metrics: {e}")
                time.sleep(5)
    
    def analyze_trends(self, entity_id: str, metric_name: str, hours_back: int = 24) -> Dict:
        """Analyze metric trends using EntityDB temporal queries"""
        try:
            # Get historical data
            history = self.client.get_entity_history(entity_id, hours_back)
            
            # Extract metric values from temporal tags
            values = []
            for entry in history:
                if entry.get("type") == "tag_change" and "value:" in entry.get("new_value", ""):
                    try:
                        value = float(entry["new_value"].split("value:")[1])
                        timestamp = entry["timestamp"]
                        values.append({"timestamp": timestamp, "value": value})
                    except (ValueError, IndexError):
                        continue
            
            if not values:
                return {"status": "no_data"}
            
            # Sort by timestamp
            values.sort(key=lambda x: x["timestamp"])
            
            # Calculate trend statistics
            recent_values = [v["value"] for v in values[-10:]]  # Last 10 values
            historical_values = [v["value"] for v in values[:-10]] if len(values) > 10 else recent_values
            
            current_avg = sum(recent_values) / len(recent_values)
            historical_avg = sum(historical_values) / len(historical_values)
            
            trend_direction = "increasing" if current_avg > historical_avg * 1.1 else "decreasing" if current_avg < historical_avg * 0.9 else "stable"
            
            return {
                "status": "success",
                "current_average": round(current_avg, 2),
                "historical_average": round(historical_avg, 2),
                "trend_direction": trend_direction,
                "data_points": len(values),
                "latest_value": recent_values[-1] if recent_values else 0
            }
            
        except Exception as e:
            return {"status": "error", "message": str(e)}
    
    def check_alerts(self):
        """Intelligent alerting based on historical patterns"""
        print("🚨 Checking for alerts...")
        
        alerts = []
        
        for hostname, server_id in self.servers.items():
            # Check CPU usage trends
            cpu_trend = self.analyze_trends(server_id, "cpu_usage", 2)  # 2 hours of data
            if cpu_trend.get("status") == "success":
                current_cpu = cpu_trend.get("current_average", 0)
                if current_cpu > 80:
                    alerts.append({
                        "severity": "critical" if current_cpu > 90 else "warning",
                        "type": "high_cpu_usage",
                        "target": hostname,
                        "message": f"CPU usage {current_cpu}% (trend: {cpu_trend.get('trend_direction')})",
                        "timestamp": datetime.now().isoformat()
                    })
            
            # Check memory usage trends
            memory_trend = self.analyze_trends(server_id, "memory_usage", 2)
            if memory_trend.get("status") == "success":
                current_memory = memory_trend.get("current_average", 0)
                if current_memory > 85:
                    alerts.append({
                        "severity": "warning",
                        "type": "high_memory_usage", 
                        "target": hostname,
                        "message": f"Memory usage {current_memory}% (trend: {memory_trend.get('trend_direction')})",
                        "timestamp": datetime.now().isoformat()
                    })
        
        for service_name, service_id in self.services.items():
            # Check response time trends
            response_trend = self.analyze_trends(service_id, "response_time", 1)  # 1 hour of data
            if response_trend.get("status") == "success":
                current_response = response_trend.get("current_average", 0)
                if current_response > 200:  # 200ms threshold
                    alerts.append({
                        "severity": "warning",
                        "type": "slow_response_time",
                        "target": service_name,
                        "message": f"Response time {current_response}ms (trend: {response_trend.get('trend_direction')})",
                        "timestamp": datetime.now().isoformat()
                    })
            
            # Check error rate
            error_trend = self.analyze_trends(service_id, "error_rate", 1)
            if error_trend.get("status") == "success":
                current_error_rate = error_trend.get("current_average", 0)
                if current_error_rate > 2.0:  # 2% error rate threshold
                    alerts.append({
                        "severity": "critical" if current_error_rate > 5.0 else "warning",
                        "type": "high_error_rate",
                        "target": service_name,
                        "message": f"Error rate {current_error_rate}% (trend: {error_trend.get('trend_direction')})",
                        "timestamp": datetime.now().isoformat()
                    })
        
        # Store alerts as entities for historical tracking
        for alert in alerts:
            try:
                alert_id = self.client.create_entity(
                    entity_type="alert",
                    dataset="monitoring",
                    tags={
                        "severity": alert["severity"],
                        "alert_type": alert["type"],
                        "target": alert["target"],
                        "status": "active"
                    },
                    content=json.dumps(alert)
                )
                print(f"  🚨 {alert['severity'].upper()}: {alert['message']}")
            except Exception as e:
                print(f"  ❌ Failed to store alert: {e}")
        
        if not alerts:
            print("  ✅ All systems nominal")
        
        return alerts
    
    def generate_dashboard_data(self) -> Dict:
        """Generate dashboard data showcasing temporal capabilities"""
        dashboard = {
            "timestamp": datetime.now().isoformat(),
            "servers": {},
            "services": {},
            "alerts": [],
            "system_overview": {
                "total_servers": len(self.servers),
                "total_services": len(self.services),
                "healthy_servers": 0,
                "healthy_services": 0
            }
        }
        
        # Get current server status with trends
        for hostname, server_id in self.servers.items():
            cpu_trend = self.analyze_trends(server_id, "cpu_usage", 1)
            memory_trend = self.analyze_trends(server_id, "memory_usage", 1)
            
            server_healthy = True
            if cpu_trend.get("current_average", 0) > 80 or memory_trend.get("current_average", 0) > 85:
                server_healthy = False
            
            if server_healthy:
                dashboard["system_overview"]["healthy_servers"] += 1
            
            dashboard["servers"][hostname] = {
                "status": "healthy" if server_healthy else "warning",
                "cpu": cpu_trend,
                "memory": memory_trend
            }
        
        # Get current service status with trends
        for service_name, service_id in self.services.items():
            response_trend = self.analyze_trends(service_id, "response_time", 1)
            error_trend = self.analyze_trends(service_id, "error_rate", 1)
            
            service_healthy = True
            if response_trend.get("current_average", 0) > 200 or error_trend.get("current_average", 0) > 2.0:
                service_healthy = False
            
            if service_healthy:
                dashboard["system_overview"]["healthy_services"] += 1
            
            dashboard["services"][service_name] = {
                "status": "healthy" if service_healthy else "warning",
                "response_time": response_trend,
                "error_rate": error_trend
            }
        
        # Get recent alerts
        recent_alerts = self.client.query_entities("type:alert")
        dashboard["alerts"] = recent_alerts[:10]  # Last 10 alerts
        
        return dashboard
    
    def start_monitoring(self):
        """Start the monitoring system"""
        print("🚀 Starting EntityDB Monitoring System...")
        print("=" * 60)
        
        self.running = True
        
        # Start metric collection in background thread
        collector_thread = threading.Thread(target=self.collect_metrics, daemon=True)
        collector_thread.start()
        
        # Main monitoring loop
        try:
            while self.running:
                print(f"\n📊 MONITORING DASHBOARD - {datetime.now().strftime('%H:%M:%S')}")
                print("-" * 60)
                
                # Check for alerts
                self.check_alerts()
                
                # Generate and display dashboard summary
                dashboard = self.generate_dashboard_data()
                overview = dashboard["system_overview"]
                
                print(f"\n🎯 SYSTEM OVERVIEW:")
                print(f"  Servers: {overview['healthy_servers']}/{overview['total_servers']} healthy")
                print(f"  Services: {overview['healthy_services']}/{overview['total_services']} healthy")
                print(f"  Active Alerts: {len(dashboard['alerts'])}")
                
                # Show sample server trends
                print(f"\n🖥️  SERVER TRENDS (sample):")
                for hostname, data in list(dashboard["servers"].items())[:3]:
                    cpu_avg = data["cpu"].get("current_average", 0)
                    memory_avg = data["memory"].get("current_average", 0)
                    print(f"  {hostname}: CPU {cpu_avg}%, Memory {memory_avg}% [{data['status']}]")
                
                print(f"\n⏱️  Demonstrating EntityDB temporal capabilities:")
                print(f"  - Collecting metrics every 10 seconds with nanosecond precision")
                print(f"  - Historical trend analysis using temporal queries")
                print(f"  - Intelligent alerting based on historical patterns")
                print(f"  - Point-in-time recovery and as-of queries available")
                
                time.sleep(30)  # Update dashboard every 30 seconds
                
        except KeyboardInterrupt:
            print(f"\n⏹️  Monitoring stopped by user")
        finally:
            self.running = False
    
    def stop_monitoring(self):
        """Stop the monitoring system"""
        self.running = False
        print("🛑 Monitoring system stopped")

# Need to import math for sin function
import math

if __name__ == "__main__":
    print("🎯 EntityDB Temporal Monitoring System Demo")
    print("=" * 50)
    print("This demonstration showcases EntityDB's temporal database capabilities")
    print("with a real-world monitoring system that tracks servers and services.")
    print("")
    print("Features demonstrated:")
    print("• Nanosecond-precision temporal data storage")
    print("• Historical trend analysis using temporal queries")  
    print("• Point-in-time queries (as-of functionality)")
    print("• Intelligent alerting based on historical patterns")
    print("• Real-time metric collection and analysis")
    print("")
    print("Press Ctrl+C to stop the demonstration")
    print("=" * 50)
    
    try:
        monitoring = MonitoringSystem()
        monitoring.start_monitoring()
    except KeyboardInterrupt:
        print("\n👋 Demo completed!")
    except Exception as e:
        print(f"\n❌ Error: {e}")