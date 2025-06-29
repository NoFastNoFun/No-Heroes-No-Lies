extends Button

# CardWidget - Individual card display component
# Used in the CardBrowser to show cards in a grid layout

signal card_clicked(card_data: Dictionary)

# UI References
@onready var card_image: TextureRect = $VBoxContainer/CardImage
@onready var card_name: Label = $VBoxContainer/CardName
@onready var card_type: Label = $VBoxContainer/CardType
@onready var card_strength: Label = $VBoxContainer/CardStrength

# Card data
var card_data: Dictionary = {}

func _ready():
	# Connect the pressed signal
	pressed.connect(_on_pressed)

func setup_card(data: Dictionary):
	card_data = data
	
	# Update UI elements
	card_name.text = data.name
	card_type.text = data.type.capitalize()
	card_strength.text = "Strength: " + str(data.strength)
	
	# Set card type color
	if data.type == "hero":
		card_type.add_theme_color_override("font_color", Color(0.2, 0.8, 0.2))  # Green for heroes
	else:
		card_type.add_theme_color_override("font_color", Color(0.8, 0.2, 0.2))  # Red for monsters
	
	# Load card image if available
	if data.has("cover_url") and data.cover_url != "":
		load_card_image(data.cover_url)
	else:
		create_placeholder_image()

func load_card_image(image_url: String):
	# In a real implementation, you'd load the image from the URL
	# For now, we'll create a placeholder based on card type
	create_placeholder_image()

func create_placeholder_image():
	var image = Image.create(100, 120, false, Image.FORMAT_RGBA8)
	
	# Create different colored placeholders for different card types
	if card_data.type == "hero":
		image.fill(Color(0.2, 0.4, 0.2, 1.0))  # Dark green for heroes
	else:
		image.fill(Color(0.4, 0.2, 0.2, 1.0))  # Dark red for monsters
	
	# Add some visual elements to make it look like a card
	var texture = ImageTexture.create_from_image(image)
	card_image.texture = texture

func _on_pressed():
	card_clicked.emit(card_data)

# Public methods
func get_card_data() -> Dictionary:
	return card_data

func update_strength_display(new_strength: int):
	card_strength.text = "Strength: " + str(new_strength)

func highlight_card():
	# Add visual feedback for highlighting
	modulate = Color(1.2, 1.2, 1.2, 1.0)

func unhighlight_card():
	# Remove highlighting
	modulate = Color(1.0, 1.0, 1.0, 1.0)

func set_card_enabled(enabled: bool):
	disabled = not enabled
	if not enabled:
		modulate = Color(0.5, 0.5, 0.5, 1.0)
	else:
		modulate = Color(1.0, 1.0, 1.0, 1.0) 