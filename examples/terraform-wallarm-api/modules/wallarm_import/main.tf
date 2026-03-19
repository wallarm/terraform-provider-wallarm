# ─── Import existing rules ─────────────────────────────────────────────────────

module "import_rules" {
  source       = "./modules/import_rules"
  # is_importing = var.is_importing
  is_importing = false
  client_id    = var.client_id
}

# ─── Import existing applications ─────────────────────────────────────────────

module "import_applications" {
  source       = "./modules/import_applications"
  # is_importing = var.is_importing
  is_importing = false
  client_id    = var.client_id
}

# ─── Import existing IP lists ────────────────────────────────────────────────

module "import_ip_lists" {
  source             = "./modules/import_ip_lists"
  is_importing       = var.is_importing
  client_id          = var.client_id
  subnet_import_mode = var.subnet_import_mode
}
