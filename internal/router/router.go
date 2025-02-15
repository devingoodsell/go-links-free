// Add bulk operation routes
func (r *Router) setupLinkRoutes() {
    // ... existing routes ...
    r.router.HandleFunc("/api/admin/links/bulk-delete", r.linkHandler.BulkDelete).Methods("POST")
    r.router.HandleFunc("/api/admin/links/bulk-status", r.linkHandler.BulkUpdateStatus).Methods("POST")
} 