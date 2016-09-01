// Copyright 2016 Tera Insights, LLC. All Rights Reserved.

package server

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

// DBHandler abstracts a variety of CRUD operations.
type DBHandler struct {
	DB     *gorm.DB
	Writer http.ResponseWriter
}

// CountWhere returns the number of records that satisfy the query `q`.
func (h *DBHandler) CountWhere(q interface{}) int {
	c := 0
	h.DB.Where(q).Count(&c)
	return c
}

// FirstWhere writes, into `r`, the first record that satisfies `q`.
func (h *DBHandler) FirstWhere(q interface{}, r interface{}) {
	h.DB.Where(q).First(r)
}

// FirstWhereWithRespond finds the first record that satisfies `q` and writes
// it to `h.Writer`.
func (h *DBHandler) FirstWhereWithRespond(q interface{}, r interface{}) {
	count := h.CountWhere(q)
	if count > 0 {
		h.FirstWhere(q, r)
		writeJSON(h.Writer, http.StatusOK, r)
	} else {
		http.Error(h.Writer, "Could not find resource", http.StatusNotFound)
	}
}
