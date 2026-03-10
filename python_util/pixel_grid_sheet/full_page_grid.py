from PIL import Image, ImageDraw

# Letter page 300 DPI
WIDTH = 2550
HEIGHT = 3300

CELL = 75         # math exact
GRIDLINE = 2
OUTER = 8
MARGIN = 150      # 1/2 inch borders

img = Image.new("RGB", (WIDTH, HEIGHT), "white")
draw = ImageDraw.Draw(img)

# Compute number of full cells that fit inside the margins
cols = (WIDTH - 2 * MARGIN) // CELL
rows = (HEIGHT - 2 * MARGIN) // CELL

grid_width = cols * CELL
grid_height = rows * CELL

# Center grid within margins
offset_x = MARGIN + (WIDTH - 2 * MARGIN - grid_width) // 2
offset_y = MARGIN + (HEIGHT - 2 * MARGIN - grid_height) // 2

# Draw vertical lines
for c in range(cols + 1):
    x = offset_x + c * CELL
    draw.line((x, offset_y, x, offset_y + grid_height), fill="black", width=GRIDLINE)

# Draw horizontal lines
for r in range(rows + 1):
    y = offset_y + r * CELL
    draw.line((offset_x, y, offset_x + grid_width, y), fill="black", width=GRIDLINE)

# Outer thick border
draw.rectangle(
    (offset_x, offset_y, offset_x + grid_width, offset_y + grid_height),
    outline="black",
    width=OUTER
)

# Save PDF
img.save("pixel_art_big_grid_75px.pdf", "PDF", resolution=300)

print("pixel_art_big_grid_75px.pdf created")