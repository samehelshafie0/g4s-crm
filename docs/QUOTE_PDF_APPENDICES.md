# Quote PDFs with supporting documents

Implemented 2026-09-17. Supports PDF, PNG/JPG, Word (DOC/DOCX) and Excel (XLS/XLSX).

## Using it

1. Open a draft quote and select **Documents**.
2. Include files from the document library, or use **Upload and include**. The default library filter shows this customer's linked documents and documents without customer links. **All accessible documents** is available when needed.
3. Give each appendix a customer-facing label and use the up/down controls to choose its order. Remove excludes a file from this quote; it does not delete the library document.
4. Save the draft. **Export PDF**, available in the quote header on every tab, also saves an editable draft before generating the PDF. If saving fails or the quote changed concurrently, export stops.
5. The downloaded file contains the quotation, an index with appendix page ranges, and every selected document's pages in order. Each appendix page carries its label, version and document page number. All pages carry the quote number and continuous `Page X of Y`. PDF bookmarks provide navigation.

Selections belong to the quote and are distinct from generic document links. Selecting a file pins its immutable document version and snapshots its name. Uploading a newer library version or renaming/deleting the library entry does not silently change an existing quote's selected file. In a draft, remove and reselect to use the latest version. Revisions/duplicates retain the source selections; submitted quotes cannot change them.

New uploads remain in the document library even if the quote is not saved. Uploading does not itself approve a quote, record customer acceptance or send a message.

## Export behavior and limits

- Quote values come from the saved server record, including its currency, totals, addresses, selected printable rows, introduction, payment/delivery terms, statement of work and closing. Internal notes, purchasing notes, generic CRM notes, costs and margins are excluded.
- Native PDF content is retained with extra space for labels, preserving its visible crop and orientation. Interactive fields remain imported; cropped-out annotations are hidden. Original uploaded files are not modified. The combined document is a new PDF, not the original signed file.
- Images are fitted without stretching. Word and Excel files are converted by LibreOffice using their print layout. Spreadsheet print areas, sheet visibility and page setup determine the printed pages. Install any required company fonts on the server for consistent Office layout.
- Up to 20 selected files, 50 MB per file, 100 MB combined input and 500 pages including the quote. Two exports can run concurrently per API process, with a three-minute job deadline. Busy responses ask the user to retry.
- Missing, corrupt, unsupported or password-protected input fails the complete export with a visible error; conversion errors identify the affected appendix. Files are never silently omitted or replaced by filenames alone.
- Documents and quote permissions remain server-enforced. New selections require document-read access; exporting requires quote-read access. A selected file stays part of its readable quote even after library soft deletion.

## API and database

- Migration `000005_quote_appendices` adds an ordered quote-to-document-version relation with unique quote/version and quote/order constraints. It is applied locally; existing applied migrations were not edited.
- `GET /api/v1/quotes/:id/builder` includes `appendices` with IDs, labels, version/file metadata and order.
- `PUT /api/v1/quotes/:id/builder` accepts optional `appendices: [{documentVersionId, label}]` in export order. An omitted field preserves current selections; `[]` removes them. The quote lock/version and transaction cover both rows and appendices. Duplicate versions, invalid references and oversized selections are rejected atomically.
- `GET /api/v1/quotes/:id/pdf?lockVersion=N` returns an authenticated `application/pdf` download with `Cache-Control: no-store`. A stale version returns 409. The backend stages source files opened through the configured storage root into a private temporary job directory, then invokes the embedded Python renderer without a shell.
- Runtime dependencies are pinned in `crm-api/pdf-requirements.txt`. LibreOffice uses an isolated per-file profile with macros and automatic link updates disabled. Temporary files are removed after completion or cancellation.

Implementation references: [LibreOffice command-line conversion](https://help.libreoffice.org/latest/en-GB/text/shared/guide/start_parameters.html), [pypdf merging and rotation](https://pypdf.readthedocs.io/en/stable/user/merging-pdfs.html).

## Setup and verification

Local setup from the repository root:

```sh
make dev-pdf
# macOS; on Linux install libreoffice-writer, libreoffice-calc and fonts-dejavu-core
brew install --cask libreoffice
make dev-migrate
make dev-api
make test-pdf
make test-integration
```

The development launcher detects `crm-api/tmp/pdf-venv/bin/python`. Override `PDF_PYTHON` or `SOFFICE_PATH` if necessary. The API Dockerfile installs the Python environment, LibreOffice Writer/Calc and fonts. CI installs the same dependencies and runs renderer tests before the Go suite.

Evidence for this batch:

- 27 PostgreSQL integration groups with race detector pass, including appendix ordering/labels, stale writes, atomic rejection, older-client saves, pinned metadata, revision copying, immutable submitted quotes, authenticated export and library soft deletion.
- Five renderer tests pass: PDF/DOCX/XLSX/PNG combined export; legacy DOC/XLS; cropped content and retained form fields; quote-only/optional-row exclusion; encrypted/corrupt-file failure. A 13-page example (six quote plus seven appendix pages) was rendered and visually inspected.
- 14 frontend tests and checked build pass; Go vet passes. The existing parser bundle warning remains.
- Browser: four existing files selected, reordered, relabeled, saved and reloaded; **Export PDF** downloaded an eight-page mixed-document PDF with matching index and bookmarks. Mobile document tab fits 390px; all tabs remain visible.
- Native browser file selection was blocked by the Chrome extension's file-URL setting. Multipart upload was verified through the API; the upload form itself was inspected. This is a test-tool limitation, not an application permission requirement.
- Docker image verification is blocked: two build attempts timed out fetching base-image metadata from Docker Hub. No image/deployment success is claimed. Local conversion/runtime verification passes; repeat image build before deployment.

Local QA quote `QT-2026-0002` and four clearly named `QA` documents remain available to inspect. They are synthetic verification records. This batch has not been pushed or deployed.
