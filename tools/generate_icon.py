import os
from PIL import Image, ImageDraw

def create_app_icon():
    os.makedirs("assets", exist_ok=True)
    size = 256
    img = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    # 1. Background squircle with subtle gradient simulation
    margin = 12
    radius = 54
    # Outer glow / subtle border
    draw.rounded_rectangle(
        [margin, margin, size - margin, size - margin],
        radius=radius,
        fill=(15, 23, 42, 255),       # Slate 900
        outline=(56, 189, 248, 200),  # Sky blue neon border
        width=4
    )

    # Inner soft background plate
    inner_m = margin + 5
    draw.rounded_rectangle(
        [inner_m, inner_m, size - inner_m, size - inner_m],
        radius=radius - 4,
        fill=(13, 17, 23, 255)         # GitHub / Dark theme black
    )

    # 2. Terminal header dots (Mac/CLI window bar style)
    dot_y = 38
    draw.ellipse([36, dot_y, 48, dot_y + 12], fill=(239, 68, 68, 220))   # Red dot
    draw.ellipse([56, dot_y, 68, dot_y + 12], fill=(245, 158, 11, 220))  # Amber dot
    draw.ellipse([76, dot_y, 88, dot_y + 12], fill=(16, 185, 129, 220))  # Emerald dot

    # Header separator subtle line
    draw.line([(32, 60), (size - 32, 60)], fill=(30, 41, 59, 255), width=2)

    # 3. Main Claude Terminal Prompt: ">" symbol (Amber/Orange Claude style)
    # Chevron points: (52, 100) -> (96, 140) -> (52, 180)
    prompt_color = (245, 158, 11, 255) # Amber 500
    p1 = (50, 95)
    p2 = (96, 140)
    p3 = (50, 185)
    draw.line([p1, p2], fill=prompt_color, width=14, joint="curve")
    draw.line([p2, p3], fill=prompt_color, width=14, joint="curve")

    # 4. Session stream & compression blocks (Cyan / Sky blue)
    # Block 1 (Top full stream)
    draw.rounded_rectangle([122, 100, 208, 116], radius=6, fill=(56, 189, 248, 255))
    # Block 2 (Mid compacted stream)
    draw.rounded_rectangle([122, 132, 185, 148], radius=6, fill=(14, 165, 233, 220))
    # Block 3 (Bottom compact summary / handoff chip)
    draw.rounded_rectangle([122, 164, 160, 180], radius=6, fill=(2, 132, 199, 190))

    # 5. Downward compression funnel / arrow badge
    # Draw a neat lightning/arrow badge representing high compression & handoff
    arrow_color = (251, 191, 36, 255) # Bright amber
    draw.polygon([(182, 142), (206, 142), (194, 166)], fill=arrow_color)
    draw.line([(194, 126), (194, 154)], fill=arrow_color, width=6)

    # Save PNG
    png_path = "assets/icon.png"
    img.save(png_path, "PNG")

    # Save Multi-size ICO
    ico_path = "assets/icon.ico"
    sizes = [(256, 256), (128, 128), (64, 64), (48, 48), (32, 32), (16, 16)]
    img.save(ico_path, format="ICO", sizes=sizes)
    print(f"Generated {png_path} and {ico_path} successfully.")

if __name__ == "__main__":
    create_app_icon()
