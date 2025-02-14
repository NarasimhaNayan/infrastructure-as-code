variable "resource_group_name" {
  description = "Resource group name for network resources"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "virtual_network_name" {
  description = "Name of the Virtual Network"
  type        = string
}

variable "address_space" {
  description = "Address space for the Virtual Network"
  type        = list(string)
}

variable "public_subnet_prefix" {
  description = "Address prefix for the public subnet"
  type        = string
}

variable "private_subnet_prefix" {
  description = "Address prefix for the private subnet"
  type        = string
}
