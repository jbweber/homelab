#!/usr/bin/env python3
"""
Nook API Python Client Example

This example demonstrates how to interact with the Nook metadata service API
using Python requests library.
"""

import requests
import json
from typing import Dict, List, Optional

class NookClient:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'User-Agent': 'nook-python-client/1.0'
        })
    
    def get_machines(self) -> List[Dict]:
        """Get list of all machines."""
        response = self.session.get(f"{self.base_url}/api/v0/machines")
        response.raise_for_status()
        return response.json()
    
    def create_machine(self, name: str, hostname: str, ipv4: str = None, network_id: int = None) -> Dict:
        """Create a new machine."""
        data = {
            "name": name,
            "hostname": hostname
        }
        if ipv4:
            data["ipv4"] = ipv4
        if network_id:
            data["network_id"] = network_id
            
        response = self.session.post(f"{self.base_url}/api/v0/machines", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_machine_by_name(self, name: str) -> Dict:
        """Get machine by name."""
        response = self.session.get(f"{self.base_url}/api/v0/machines/name/{name}")
        response.raise_for_status()
        return response.json()
    
    def delete_machine(self, machine_id: int) -> None:
        """Delete a machine."""
        response = self.session.delete(f"{self.base_url}/api/v0/machines/{machine_id}")
        response.raise_for_status()
    
    def get_networks(self) -> List[Dict]:
        """Get list of all networks."""
        response = self.session.get(f"{self.base_url}/api/v0/networks")
        response.raise_for_status()
        return response.json()
    
    def create_network(self, name: str, bridge: str, subnet: str, 
                      gateway: str = None, dns_servers: str = None, 
                      description: str = None) -> Dict:
        """Create a new network."""
        data = {
            "name": name,
            "bridge": bridge,
            "subnet": subnet
        }
        if gateway:
            data["gateway"] = gateway
        if dns_servers:
            data["dns_servers"] = dns_servers
        if description:
            data["description"] = description
            
        response = self.session.post(f"{self.base_url}/api/v0/networks", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_ssh_keys(self) -> List[Dict]:
        """Get list of all SSH keys."""
        response = self.session.get(f"{self.base_url}/api/v0/ssh-keys")
        response.raise_for_status()
        return response.json()
    
    def add_ssh_key(self, machine_id: int, key_text: str) -> Dict:
        """Add SSH key to a machine."""
        data = {
            "machine_id": machine_id,
            "key_text": key_text
        }
        response = self.session.post(f"{self.base_url}/api/v0/ssh-keys", json=data)
        response.raise_for_status()
        return response.json()
    
    def delete_ssh_key(self, key_id: int) -> None:
        """Delete an SSH key."""
        response = self.session.delete(f"{self.base_url}/api/v0/ssh-keys/{key_id}")
        response.raise_for_status()


def example_usage():
    """Example usage of the Nook API client."""
    client = NookClient("http://localhost:8080")
    
    try:
        # Create a network
        print("Creating network...")
        network = client.create_network(
            name="example-net",
            bridge="br0", 
            subnet="192.168.100.0/24",
            gateway="192.168.100.1",
            description="Example network for testing"
        )
        print(f"Created network: {network['name']} (ID: {network['id']})")
        
        # Create a machine
        print("Creating machine...")
        machine = client.create_machine(
            name="example-vm",
            hostname="example-vm.local",
            ipv4="192.168.100.10",
            network_id=network['id']
        )
        print(f"Created machine: {machine['name']} (ID: {machine['id']})")
        
        # Add SSH key
        print("Adding SSH key...")
        with open('/home/user/.ssh/id_rsa.pub', 'r') as f:
            ssh_key_text = f.read().strip()
        
        ssh_key = client.add_ssh_key(machine['id'], ssh_key_text)
        print(f"Added SSH key (ID: {ssh_key['id']})")
        
        # List all resources
        print("Listing all machines:")
        machines = client.get_machines()
        for m in machines:
            print(f"  - {m['name']} ({m['ipv4']})")
        
        print("Listing all networks:")
        networks = client.get_networks()
        for n in networks:
            print(f"  - {n['name']} ({n['subnet']})")
            
        print("Listing all SSH keys:")
        keys = client.get_ssh_keys()
        for k in keys:
            print(f"  - Key {k['id']} for machine {k['machine_id']}")
            
    except requests.exceptions.RequestException as e:
        print(f"API request failed: {e}")
    except FileNotFoundError:
        print("SSH key file not found. Update the path in the example.")
    except Exception as e:
        print(f"Unexpected error: {e}")


if __name__ == "__main__":
    example_usage()