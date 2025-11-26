package models

// SignupRequest represents the user signup request body
type SignupRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100" example:"John"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100" example:"Doe"`
	Email     string `json:"email" validate:"email,required" example:"john.doe@example.com"`
	Password  string `json:"password" validate:"required,min=6" example:"password123"`
	Phone     string `json:"phone" validate:"required" example:"+1234567890"`
	UserType  string `json:"user_type" validate:"required,eq=ADMIN|eq=USER" example:"USER" enums:"USER,ADMIN"`
}

// SignupResponse represents the successful signup response
type SignupResponse struct {
	Message string `json:"message" example:"User created successfully"`
	Token   string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    User   `json:"user"`
}

// LoginResponse represents the successful login response
type LoginResponse struct {
	Message string      `json:"message" example:"Login successful"`
	Token   string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    UserSummary `json:"user"`
}

// UserSummary represents basic user info in responses
type UserSummary struct {
	ID        string `json:"id" example:"507f1f77bcf86cd799439011"`
	Email     string `json:"email" example:"john.doe@example.com"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	UserType  string `json:"user_type" example:"USER" enums:"USER,ADMIN"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error message here"`
}

// SuccessResponse represents a generic success message
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// FoodCreateRequest represents the request to create a food item
type FoodCreateRequest struct {
	Name      string  `json:"name" validate:"required,min=2,max=100" example:"Grilled Chicken"`
	Price     float64 `json:"price" validate:"required,gt=0" example:"15.99"`
	FoodImage string  `json:"food_image" validate:"required" example:"https://example.com/images/chicken.jpg"`
	MenuID    string  `json:"menu_id" validate:"required" example:"507f1f77bcf86cd799439011"`
}

// FoodResponse represents the response after creating a food item
type FoodResponse struct {
	Message string `json:"message" example:"Food created successfully"`
	ID      string `json:"id" example:"507f1f77bcf86cd799439011"`
	Food    Food   `json:"food"`
}

// MenuCreateRequest represents the request to create a menu
type MenuCreateRequest struct {
	Name      string `json:"name" validate:"required" example:"Dinner Menu"`
	Category  string `json:"category" validate:"required" example:"Main Course"`
	StartDate string `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate   string `json:"end_date" example:"2024-12-31T23:59:59Z"`
}

// OrderCreateRequest represents the request to create an order
type OrderCreateRequest struct {
	TableID string `json:"table_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	Status  string `json:"status" validate:"required" example:"pending" enums:"pending,preparing,ready,delivered,cancelled"`
}

// TableCreateRequest represents the request to create a table
type TableCreateRequest struct {
	TableNumber int `json:"table_number" validate:"required,min=1" example:"5"`
	Capacity    int `json:"capacity" validate:"required,min=1" example:"4"`
}

// InvoiceCreateRequest represents the request to create an invoice
type InvoiceCreateRequest struct {
	OrderID       string  `json:"order_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	PaymentMethod string  `json:"payment_method" validate:"required" example:"credit_card" enums:"cash,credit_card,debit_card,mobile_payment"`
	TotalAmount   float64 `json:"total_amount" validate:"required,gt=0" example:"45.99"`
	PaymentStatus string  `json:"payment_status" validate:"required" example:"paid" enums:"pending,paid,failed,refunded"`
}

// OrderItemCreateRequest represents the request to create an order item
type OrderItemCreateRequest struct {
	OrderID   string  `json:"order_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	FoodID    string  `json:"food_id" validate:"required" example:"507f1f77bcf86cd799439013"`
	Quantity  int     `json:"quantity" validate:"required,min=1" example:"2"`
	UnitPrice float64 `json:"unit_price" validate:"required,gt=0" example:"15.99"`
}

// MenuResponse represents the response after creating or fetching a menu
type MenuResponse struct {
	Message string `json:"message" example:"Menu fetched successfully"`
	ID      string `json:"id" example:"507f1f77bcf86cd799439014"`
	Menu    Menu   `json:"menu"`
}

// OrderResponse represents the response after creating or fetching an order
type OrderResponse struct {
	Message string `json:"message" example:"Order fetched successfully"`
	ID      string `json:"id" example:"507f1f77bcf86cd799439015"`
	Order   Order  `json:"order"`
}

// TableResponse represents the response after creating or fetching a table
type TableResponse struct {
	Message string `json:"message" example:"Table fetched successfully"`
	ID      string `json:"id" example:"507f1f77bcf86cd799439016"`
	Table   Table  `json:"table"`
}

// OrderItemResponse represents the response after creating or fetching an order item
type OrderItemResponse struct {
	Message   string    `json:"message" example:"Order item fetched successfully"`
	ID        string    `json:"id" example:"507f1f77bcf86cd799439017"`
	OrderItem OrderItem `json:"order_item"`
}

// InvoiceResponse represents the response after creating or fetching an invoice
type InvoiceResponse struct {
	Message string  `json:"message" example:"Invoice fetched successfully"`
	ID      string  `json:"id" example:"507f1f77bcf86cd799439018"`
	Invoice Invoice `json:"invoice"`
}
