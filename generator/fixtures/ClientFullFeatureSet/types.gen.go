package generated

// ListItemsParams is generated for operationId ListItems.
type ListItemsParams struct {
	Page     *float64
	PageSize int64
}

// Item is generated from components.schemas.Item.
type Item struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ListItemsResponse200 struct {
	Items []Item `json:"items,omitempty"`
	Total int64  `json:"total,omitempty"`
}

type CreateItemRequest struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
}

type CreateItemResponse201 struct {
	Item
	CreatedAt DateTime `json:"createdAt,omitzero"`
}

type Response400 struct{}

func (r Response400) Error() string {
	panic("TODO: define the output")
}

type Response401 struct{}

func (r Response401) Error() string {
	panic("TODO: define the output")
}

// GetItemParams is generated for operationId GetItem.
type GetItemParams struct {
	XApiKey    string
	XRequestId *string
	ItemId     string
}

// ValidationError is generated from components.schemas.ValidationError.
type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

func (r ValidationError) Error() string {
	panic("TODO: define the output")
}

type Response404 struct{}

func (r Response404) Error() string {
	panic("TODO: define the output")
}

// ReplaceItemParams is generated for operationId ReplaceItem.
type ReplaceItemParams struct {
	ItemId string
}

type ReplaceItemRequest struct {
	Item
}

type ReplaceItemResponse200 struct {
	Id        string   `json:"id,omitempty"`
	UpdatedAt DateTime `json:"updatedAt,omitzero"`
}

// ConflictError is generated from components.schemas.ConflictError.
type ConflictError struct {
	Message string `json:"message,omitempty"`
}

func (r ConflictError) Error() string {
	panic("TODO: define the output")
}

// DeleteItemParams is generated for operationId DeleteItem.
type DeleteItemParams struct {
	ItemId string
}

// NotFoundError is generated from components.schemas.NotFoundError.
type NotFoundError struct {
	ResourceId string `json:"resourceId,omitempty"`
}

func (r NotFoundError) Error() string {
	panic("TODO: define the output")
}

// ArchiveItemParams is generated for operationId ArchiveItem.
type ArchiveItemParams struct {
	ItemId string
}
