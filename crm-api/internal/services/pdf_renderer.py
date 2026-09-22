"""Quote PDF renderer. Only invoked with a server-created job in a private directory."""
import io
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path
from xml.sax.saxutils import escape

from PIL import Image, ImageOps
from pypdf import PdfReader, PdfWriter
from pypdf.generic import RectangleObject, ContentStream, FloatObject, NameObject, NumberObject
from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfgen.canvas import Canvas
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle, KeepTogether

MAX_PAGES = 500
BLUE = colors.HexColor('#2563eb')
GRAY = colors.HexColor('#475569')
FONT = 'Helvetica'
for candidate in ['/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf', '/System/Library/Fonts/Supplemental/Arial.ttf']:
    if Path(candidate).exists():
        pdfmetrics.registerFont(TTFont('QuoteText', candidate))
        FONT = 'QuoteText'
        break
STYLE = ParagraphStyle('Body', fontName=FONT, fontSize=9, leading=13, textColor=colors.HexColor('#1e293b'), spaceAfter=6, splitLongWords=True)
SMALL = ParagraphStyle('Small', parent=STYLE, fontSize=8, leading=11)
HEADING = ParagraphStyle('Heading', parent=STYLE, fontSize=13, leading=18, spaceBefore=12, spaceAfter=8, textColor=BLUE)


def para(value, style=STYLE):
    return Paragraph(escape(str(value or '')).replace('\n', '<br/>'), style)


def money(value):
    return f'{float(value or 0):,.2f}'


def address(value, fallback=''):
    return '\n'.join(str(v) for v in [value.get('company') or fallback, value.get('contactName'), value.get('address'), ', '.join(v for v in [value.get('city'), value.get('country')] if v), value.get('phone'), value.get('email')] if v)


def quote_pdf(q, appendices, quote_pages=0):
    output = io.BytesIO()
    doc = SimpleDocTemplate(output, pagesize=A4, leftMargin=36, rightMargin=36, topMargin=42, bottomMargin=42, title=q['quoteNumber'], author='G4S CRM')
    story = [para('G4S | Security Solutions', HEADING), para('Quotation ' + q['quoteNumber'], HEADING)]
    story.append(para(f"Status: {q['status']}   |   Currency: {q['currency']}   |   Valid until: {(q.get('validUntil') or '')[:10] or '-'}", SMALL))
    sold = q.get('soldTo') or {}
    ship = q.get('shipTo') or {}
    addresses = [[para('Sold to', HEADING), para('Ship to', HEADING)], [para(address(sold, (q.get('customer') or {}).get('companyName', ''))), para(address(ship))]]
    story += [Table(addresses, colWidths=[261.6, 261.6], style=[('VALIGN', (0,0), (-1,-1), 'TOP'), ('LEFTPADDING',(0,0),(-1,-1),0)]), Spacer(1, 14)]
    if q.get('introductionText'): story.append(para(q['introductionText']))
    data = [[para(v, SMALL) for v in ['#', 'Description', 'Qty × factor', 'Unit price', 'Discount', 'Total']]]
    spans = []
    running, item_index = 0, 0
    for line in q.get('lineItems') or []:
        if line.get('isOptional') and not line.get('isSelected'): continue
        kind = line.get('rowType') or 'item'
        if kind == 'item': running += float(line.get('lineTotal') or 0)
        if not line.get('isPrintable', True): continue
        if kind == 'item':
            item_index += 1
            description = line.get('description', '')
            if line.get('sku'): description += '\n' + line['sku']
            qty = f"{line['quantity']:g} × {line.get('multiplier', 1):g}"
            data.append([para(item_index, SMALL), para(description), para(qty, SMALL), para(money(line.get('unitPrice')), SMALL), para(f"{line.get('discountPercent', 0):g}%", SMALL), para(money(line.get('lineTotal')), SMALL)])
        else:
            text = line.get('headingText') if kind == 'heading' else line.get('commentText') if kind == 'comment' else f"Subtotal: {q['currency']} {money(running)}" if kind == 'subtotal' else line.get('description', '')
            data.append([para(text, HEADING if kind == 'heading' else STYLE), '', '', '', '', ''])
            spans.append(('SPAN', (0,len(data)-1), (-1,len(data)-1)))
    table = Table(data, colWidths=[23, 193, 65, 80, 62, 100.2], repeatRows=1, splitByRow=1, splitInRow=1)
    table.setStyle(TableStyle([('VALIGN',(0,0),(-1,-1),'TOP'), ('BACKGROUND',(0,0),(-1,0),colors.HexColor('#eff6ff')), ('LINEBELOW',(0,0),(-1,0),1,BLUE), ('LINEBELOW',(0,1),(-1,-1),0.3,colors.HexColor('#e2e8f0')), ('TOPPADDING',(0,0),(-1,-1),7), ('BOTTOMPADDING',(0,0),(-1,-1),7)] + spans))
    story += [table, Spacer(1,16)]
    totals = [('Subtotal', q['subtotal']), (f"Discount ({q['discountPercent']:g}%)", -q['discountAmount']), ('After discount', q['subtotalAfterDiscount']), (f"VAT ({q['vatPercent']:g}%)", q['vatAmount']), ('Grand total', q['total'])]
    totals_table = Table([[para(k), para(q['currency'] + ' ' + money(v))] for k,v in totals], colWidths=[170,140], hAlign='RIGHT')
    totals_table.setStyle(TableStyle([('LINEABOVE',(0,-1),(-1,-1),1,BLUE),('BACKGROUND',(0,-1),(-1,-1),colors.HexColor('#eff6ff')),('TOPPADDING',(0,0),(-1,-1),6),('BOTTOMPADDING',(0,0),(-1,-1),6)]))
    story.append(KeepTogether([totals_table]))
    for key, title in [('paymentTerms','Payment terms'),('deliveryTerms','Delivery terms'),('statementOfWork','Statement of work'),('closingText','Closing')]:
        if q.get(key): story.extend([para(title, HEADING), para(q[key])])
    if appendices:
        story.append(para('Included documents', HEADING))
        index_rows = [[para(v, SMALL) for v in ['Appendix / document', 'Version', 'PDF pages']]]
        start = quote_pages + 1
        for i, item in enumerate(appendices, 1):
            end = start + item['pages'] - 1
            index_rows.append([para(f"Appendix {i}: {item['label']}\n{item['fileName']}"), para(item['version']), para(f'{start}-{end}' if quote_pages else 'Pending')])
            start = end + 1
        index_table = Table(index_rows, colWidths=[365.2,60,98], repeatRows=1, splitInRow=1)
        index_table.setStyle(TableStyle([('VALIGN',(0,0),(-1,-1),'TOP'),('BACKGROUND',(0,0),(-1,0),colors.HexColor('#eff6ff')),('TOPPADDING',(0,0),(-1,-1),6),('BOTTOMPADDING',(0,0),(-1,-1),6)]))
        story.append(index_table)
    doc.build(story)
    return output.getvalue()


def convert(item, directory, index):
    path = Path(item['path'])
    ext = Path(item['fileName']).suffix.lower()
    if ext == '.pdf': return path
    target = directory / f'converted-{index}.pdf'
    if ext in ['.png', '.jpg', '.jpeg']:
        with Image.open(path) as image:
            image = ImageOps.exif_transpose(image)
            width, height = A4 if image.width <= image.height else tuple(reversed(A4))
            canvas = Canvas(str(target), pagesize=(width, height))
            from reportlab.lib.utils import ImageReader
            ratio = min((width-40)/image.width, (height-40)/image.height)
            w,h = image.width*ratio,image.height*ratio
            canvas.drawImage(ImageReader(image), (width-w)/2, (height-h)/2, w,h, mask='auto')
            canvas.save()
        return target
    if ext not in ['.doc', '.docx', '.xls', '.xlsx']:
        raise ValueError('Unsupported document format. Use PDF, PNG, JPG, DOC, DOCX, XLS or XLSX.')
    office = os.getenv('SOFFICE_PATH') or shutil.which('soffice') or shutil.which('libreoffice')
    if not office and Path('/Applications/LibreOffice.app/Contents/MacOS/soffice').exists(): office = '/Applications/LibreOffice.app/Contents/MacOS/soffice'
    if not office: raise ValueError('Word/Excel conversion is unavailable. Install LibreOffice on the server.')
    profile = directory / f'office-profile-{index}'
    (profile / 'user').mkdir(parents=True)
    # Each job has an isolated profile; never run document macros or update links.
    (profile / 'user/registrymodifications.xcu').write_text('''<?xml version="1.0" encoding="UTF-8"?><oor:items xmlns:oor="http://openoffice.org/2001/registry"><item oor:path="/org.openoffice.Office.Common/Security/Scripting"><prop oor:name="MacroSecurityLevel" oor:op="fuse"><value>3</value></prop></item><item oor:path="/org.openoffice.Office.Calc/Content/Update"><prop oor:name="Link" oor:op="fuse"><value>2</value></prop></item><item oor:path="/org.openoffice.Office.Writer/Content/Update"><prop oor:name="Link" oor:op="fuse"><value>2</value></prop></item></oor:items>''')
    source = directory / f'office-{index}{ext}'
    shutil.copyfile(path, source)
    result = subprocess.run([office, '-env:UserInstallation='+profile.as_uri(), '--headless', '--nologo', '--nodefault', '--nofirststartwizard', '--convert-to', 'pdf', '--outdir', str(directory), str(source)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=60)
    converted = source.with_suffix('.pdf')
    if result.returncode or not converted.exists(): raise ValueError('Word/Excel conversion failed. Check that the file opens and is not password protected.')
    return converted


def label_page(page, title, detail, footer, appendix):
    page.transfer_rotation_to_content()
    box = page.cropbox
    left, bottom, right, top = map(float, [box.left, box.bottom, box.right, box.top])
    if appendix:
        # Keep the source crop as an explicit content clip before adding margins.
        # Retain the imported page/field tree so interactive form fields survive.
        contents = page.get_contents()
        if contents is not None:
            clipped = ContentStream(None, page.pdf)
            clipped.operations = [([], b'q'), ([FloatObject(left), FloatObject(bottom), FloatObject(right-left), FloatObject(top-bottom)], b're'), ([], b'W'), ([], b'n')] + contents.operations + [([], b'Q')]
            page.replace_contents(clipped)
        for reference in page.get('/Annots', []):
            annotation = reference.get_object()
            rect = annotation.get('/Rect')
            if rect and (float(rect[0]) < left or float(rect[1]) < bottom or float(rect[2]) > right or float(rect[3]) > top):
                # A cropped-out widget must not reappear in the new label margins.
                annotation[NameObject('/F')] = NumberObject(int(annotation.get('/F', 0)) | 2)
        bottom -= 30
        top += 62
        page.mediabox = RectangleObject((left,bottom,right,top))
        page.cropbox = RectangleObject((left,bottom,right,top))
    overlay = io.BytesIO()
    canvas = Canvas(overlay, pagesize=(right-left,top-bottom))
    canvas.setFillColor(GRAY)
    canvas.setFont(FONT,8)
    if appendix:
        text = para(title, ParagraphStyle('Appendix', parent=STYLE, fontSize=10, leading=12, textColor=BLUE))
        _, height = text.wrap(right-left-36, 38)
        if height > 38: raise ValueError('An appendix label is too long for its page. Shorten the label and export again.')
        text.drawOn(canvas,18,top-bottom-12-height)
        canvas.setFont(FONT,7)
        canvas.drawString(18,top-bottom-54,detail)
    canvas.drawString(18,12,footer)
    canvas.save()
    from pypdf import Transformation
    page.merge_transformed_page(PdfReader(overlay).pages[0], Transformation().translate(left,bottom))


def render(job, destination):
    directory = Path(destination).parent
    appendices = job['appendices']
    readers = []
    page_count = 0
    for i,item in enumerate(appendices):
        try:
            reader = PdfReader(convert(item,directory,i))
            if reader.is_encrypted: raise ValueError('Password-protected PDFs cannot be included. Upload an unlocked copy.')
            item['pages'] = len(reader.pages)
            if not item['pages']: raise ValueError('The document has no pages.')
            page_count += item['pages']
            if page_count > MAX_PAGES: raise ValueError('Appendices exceed the 500-page export limit.')
            readers.append(reader)
        except Exception as error:
            message = 'Word/Excel conversion timed out. Try a smaller file.' if isinstance(error, subprocess.TimeoutExpired) else str(error) if isinstance(error, ValueError) else 'File cannot be converted or read. Upload a valid copy.'
            raise ValueError(f"Appendix {i+1} ({item['label']}): {message}") from error
    q = job['quote']
    first = quote_pdf(q,appendices)
    base = PdfReader(io.BytesIO(first))
    if appendices:
        for _ in range(3):
            count = len(base.pages)
            base = PdfReader(io.BytesIO(quote_pdf(q,appendices,count)))
            if len(base.pages) == count: break
        else: raise ValueError('Unable to paginate the document index.')
    if len(base.pages) + page_count > MAX_PAGES: raise ValueError('Combined quote exceeds the 500-page export limit.')
    writer = PdfWriter()
    writer.append(base, import_outline=False)
    writer.add_outline_item('Quotation ' + q['quoteNumber'],0)
    labels = [None] * len(base.pages)
    for index,(item,reader) in enumerate(zip(appendices,readers),1):
        reader.add_form_topname(f'appendix_{index}')
        start = len(writer.pages)
        writer.append(reader, import_outline=False)
        writer.add_outline_item(f"Appendix {index}: {item['label']}",start)
        labels.extend([(index,item,n+1) for n in range(item['pages'])])
    total = len(writer.pages)
    for i,(page,info) in enumerate(zip(writer.pages,labels),1):
        title = detail = ''
        if info:
            index,item,n = info
            title = f"Appendix {index}: {item['label']}"
            detail = f"Version {item['version']} | Document page {n} of {item['pages']}"
        label_page(page,title,detail,f"{q['quoteNumber']} | Page {i} of {total}",info is not None)
    writer.add_metadata({'/Title': q['quoteNumber']+' with supporting documents','/Author':'G4S CRM'})
    with open(destination,'wb') as out: writer.write(out)


if __name__ == '__main__':
    try:
        render(json.loads(Path(sys.argv[1]).read_text()), sys.argv[2])
    except Exception as exc:
        message = str(exc) if isinstance(exc,ValueError) else 'PDF generation failed. Check the documents and try again.'
        print(json.dumps({'error': message}))
        sys.exit(1)
