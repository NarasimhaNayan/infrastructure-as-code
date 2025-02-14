// Create the Virtual Network
resource "azurerm_virtual_network" "virtual_network" {
  name                = var.virtual_network_name
  address_space       = var.address_space
  location            = var.location
  resource_group_name = var.resource_group_name
}

# Create the public subnet
resource "azurerm_subnet" "public_subnet" {
  name                 = "${var.virtual_network_name}-public-subnet"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.virtual_network.name
  address_prefixes     = [var.public_subnet_prefix]
}

# Create the private subnet
resource "azurerm_subnet" "private_subnet" {
  name                 = "${var.virtual_network_name}-private-subnet"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.virtual_network.name
  address_prefixes     = [var.private_subnet_prefix]
}

# Create NSG for public subnet with a sample rule (allow HTTP)
resource "azurerm_network_security_group" "public_nsg" {
  name                = "${var.virtual_network_name}-public-nsg"
  location            = var.location
  resource_group_name = var.resource_group_name

  security_rule {
    name                       = "AllowHTTP"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

# Create NSG for private subnet with restrictive rules
resource "azurerm_network_security_group" "private_nsg" {
  name                = "${var.virtual_network_name}-private-nsg"
  location            = var.location
  resource_group_name = var.resource_group_name

  security_rule {
    name                       = "AllowInternal"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "VirtualNetwork"
    destination_address_prefix = "VirtualNetwork"
  }
}

# Associate the public NSG with the public subnet
resource "azurerm_subnet_network_security_group_association" "public_assoc" {
  subnet_id                 = azurerm_subnet.public_subnet.id
  network_security_group_id = azurerm_network_security_group.public_nsg.id
}

# Associate the private NSG with the private subnet
resource "azurerm_subnet_network_security_group_association" "private_assoc" {
  subnet_id                 = azurerm_subnet.private_subnet.id
  network_security_group_id = azurerm_network_security_group.private_nsg.id
}
