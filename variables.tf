# // General settings
variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
}

variable "location" {
  description = "Azure region for resource deployment"
  type        = string
  default     = "East US"
}

# // Virtual Network variables
variable "virtual_network_name" {
  description = "Name of the Virtual Network"
  type        = string
}

variable "vnet_address_space" {
  description = "Address space for the Virtual Network (e.g. [\"10.0.0.0/16\"])"
  type        = list(string)
}

variable "public_subnet_prefix" {
  description = "Address prefix for the public subnet (e.g. \"10.0.1.0/24\")"
  type        = string
}

variable "private_subnet_prefix" {
  description = "Address prefix for the private subnet (e.g. \"10.0.2.0/24\")"
  type        = string
}

# // AKS Cluster variables
variable "aks_cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
}

variable "aks_dns_prefix" {
  description = "DNS prefix for the AKS cluster"
  type        = string
}

variable "aks_node_count" {
  description = "Number of nodes in the AKS cluster"
  type        = number
  default     = 1
}

# The VM should atleast have 2 CPUs and 4  cores
variable "aks_vm_size" {
  description = "Size of the virtual machines for AKS nodes"
  type        = string
  default     = "Standard_B2s"
}

# // PostgreSQL variables
variable "postgres_server_name" {
  description = "Name of the PostgreSQL server"
  type        = string
}

# The user name should not be admin
variable "postgres_admin" {
  description = "Administrator username for PostgreSQL"
  type        = string
}

# Password requirements: Must include at least one uppercase letter, one lowercase letter, one digit, and one special character (such as !, @, #, $, etc.).
# https://learn.microsoft.com/en-us/previous-versions/azure/postgresql/single-server/concepts-security
variable "postgres_admin_password" {
  description = "Administrator password for PostgreSQL"
  type        = string
  sensitive   = true
}

variable "postgres_database_name" {
  description = "Name of the PostgreSQL database"
  type        = string
}

// Storage Account variable
# Storage name has to be unique, and it should contain lowercase alphabets or numbers
variable "storage_account_name" {
  description = "Name of the Storage Account (must be globally unique)"
  type        = string
}
