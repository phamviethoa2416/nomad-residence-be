package request

type AdminLoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateAdminRequest struct {
	Email    string  `json:"email"     binding:"required,email,max=255"`
	Password string  `json:"password"  binding:"required,min=8"`
	FullName string  `json:"full_name" binding:"required,max=255"`
	Phone    *string `json:"phone"     binding:"omitempty,max=20"`
	Role     string  `json:"role"      binding:"omitempty,oneof=admin superadmin"`
}

type UpdateAdminRequest struct {
	FullName *string `json:"full_name" binding:"omitempty,max=255"`
	Phone    *string `json:"phone"     binding:"omitempty,max=20"`
	Role     *string `json:"role"      binding:"omitempty,oneof=admin superadmin"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
