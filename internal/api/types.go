package api

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectMember struct {
	ID   string `json:"id"`
	User *User  `json:"user"`
	Role string `json:"role"`
}

type Environment struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type Variable struct {
	ID            string `json:"id"`
	EnvironmentID string `json:"environment_id"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	UpdatedAt     string `json:"updated_at"`
}

type DiffEntry struct {
	Key    string `json:"key"`
	Status string `json:"status"` // same, changed, missing_in_a, missing_in_b
}

type VariableVersion struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Action    string `json:"action"`
	Actor     *User  `json:"actor"`
	CreatedAt string `json:"created_at"`
}

type ImportResult struct {
	Message string `json:"message"`
	Synced  int    `json:"synced"`
}

type DeviceCodeResult struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	ExpiresIn  int    `json:"expires_in"`
}

type DevicePollResult struct {
	Status string `json:"status"` // "pending" or empty when approved
	Token  string `json:"token"`
}
