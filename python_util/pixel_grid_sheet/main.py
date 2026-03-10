from PIL import Image, ImageDraw

WIDTH = 2550
HEIGHT = 3300

ROWS = 3
COLS = 2

GRID = 16
CELL = 55
SQUARE = GRID * CELL  # 880 exact

GAP_X = 180
GAP_Y = 220

MARGIN_X = 260
MARGIN_Y = 220

OUTER = 8
GRIDLINE = 2

img = Image.new("RGB", (WIDTH, HEIGHT), "white")
draw = ImageDraw.Draw(img)


def draw_grid(x, y):

    draw.rectangle(
        (x, y, x + SQUARE, y + SQUARE),
        outline="black",
        width=OUTER
    )

    for i in range(1, GRID):
        p = i * CELL

        draw.line((x + p, y, x + p, y + SQUARE), fill="black", width=GRIDLINE)
        draw.line((x, y + p, x + SQUARE, y + p), fill="black", width=GRIDLINE)


def draw_reference(x, y):

    draw.line((x, y, x + SQUARE, y), fill="black", width=OUTER)
    draw.line((x, y, x, y + SQUARE), fill="black", width=OUTER)


def draw_blank(x, y):

    draw.rectangle(
        (x, y, x + SQUARE, y + SQUARE),
        outline="black",
        width=OUTER
    )


def registration_mark(x, y):

    s = 40
    draw.line((x - s, y, x + s, y), fill="black", width=6)
    draw.line((x, y - s, x, y + s), fill="black", width=6)


square = 0

for r in range(ROWS):
    for c in range(COLS):

        x = MARGIN_X + c * (SQUARE + GAP_X)
        y = MARGIN_Y + r * (SQUARE + GAP_Y)

        if square == 0:
            draw_reference(x, y)

        elif square == 4:
            draw_blank(x, y)

        else:
            draw_grid(x, y)

        square += 1


registration_mark(120, HEIGHT // 2)
registration_mark(WIDTH - 120, HEIGHT // 2)

img.save("pixel_art_worksheet.pdf", "PDF", resolution=300)

print("pixel_art_worksheet.pdf created")