class_name RegisterForm
extends VBoxContainer


signal form_submitted(username: String, password: String, confirm_password: String, color: Color)
signal form_cancelled()


@onready var username_field: LineEdit = %UsernameField
@onready var password_field: LineEdit = %PasswordField
@onready var confirm_password_field: LineEdit = %ConfirmPasswordField
@onready var register_button: Button = %RegisterButton
@onready var cancel_button: Button = %CancelButton
@onready var color_picker: ColorPicker = %ColorPicker


func _ready() -> void:
	register_button.pressed.connect(_on_register_button_pressed)
	cancel_button.pressed.connect(_on_cancel_button_pressed)
	username_field.text_submitted.connect(_on_field_submitted)
	password_field.text_submitted.connect(_on_field_submitted)
	confirm_password_field.text_submitted.connect(_on_field_submitted)
	
	
func _on_register_button_pressed() -> void:
	form_submitted.emit(username_field.text, password_field.text, confirm_password_field.text, color_picker.color)


func _on_cancel_button_pressed() -> void:
	form_cancelled.emit()


func _on_field_submitted(_text: String) -> void:
	_on_register_button_pressed()
