# Pinterest Growth System

A Go application that creates Pinterest pins either individually or in batches using CSV files. See Pinterest Developer API documentation at developers.pinterest.com.

## Features

- 🔐 OAuth2 authentication with Pinterest API
- 📌 Single pin creation
- 📊 Batch pin creation from CSV files
- 🖼️ Supports JPEG and PNG images
- 🔄 Base64 image encoding
- ✅ Progress tracking and error reporting

## Usage

### Environment Variables

Set the following environment variables:

```bash
PINTEREST_APP_ID=your_app_id
PINTEREST_APP_SECRET=your_app_secret
PINTEREST_REFRESH_TOKEN=your_refresh_token
PINTEREST_BOARD_ID=your_board_id
```

### Single Pin Creation

For single pin creation, set these additional environment variables:

```bash
INPUT_FILE_PATH=path/to/your/image.jpg
INPUT_TITLE="Your Pin Title"
INPUT_DESCRIPTION="Your pin description"
INPUT_LINK="https://your-website.com"
INPUT_ALT_TEXT="Alt text for accessibility"
INPUT_SECTION_ID="optional_section_id"
INPUT_NOTE="Optional note"
```

### Batch Pin Creation from CSV

For batch processing, set:

```bash
INPUT_CSV_PATH=path/to/your/pins.csv
```

#### CSV Format

The CSV file should have the following columns (header row is optional):

| file_path | title | description | link | alt_text | section_id | note |
|-----------|-------|-------------|------|----------|------------|------|
| images/photo1.jpg | Beautiful Sunset | A stunning sunset | https://example.com | Sunset over mountains | | Amazing view |
| images/photo2.png | Nature Walk | Forest path | https://example.com/nature | Forest path | section123 | Peaceful walk |

**Required Column:**
- `file_path`: Path to the image file

**Optional Columns:**
- `title`: Pin title
- `description`: Pin description  
- `link`: Website URL to link to
- `alt_text`: Accessibility alt text
- `section_id`: Pinterest board section ID
- `note`: Additional note

## Running the Application

```bash
go run main.go
```

## CSV Example

See `sample_pins.csv` for an example CSV file format.

## Error Handling

- Individual pin failures in batch mode won't stop processing
- Detailed error messages for debugging
- Progress tracking with success/failure counts
- Supports both header and no-header CSV files
