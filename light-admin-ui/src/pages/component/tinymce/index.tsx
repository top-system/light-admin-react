/**
 * /component/tinymce — TinyMCE 7 self-hosted demo (no cloud API key).
 *
 * We bundle the core, theme, icons, skin, plus a common plugin set. Images
 * picked from disk are uploaded via POST /files (our file service) and the
 * returned URL is inserted back into the editor.
 */
import 'tinymce/tinymce';
import 'tinymce/models/dom';
import 'tinymce/themes/silver';
import 'tinymce/icons/default';
import 'tinymce/skins/ui/oxide/skin.js';
import 'tinymce/plugins/advlist';
import 'tinymce/plugins/autolink';
import 'tinymce/plugins/code';
import 'tinymce/plugins/fullscreen';
import 'tinymce/plugins/image';
import 'tinymce/plugins/link';
import 'tinymce/plugins/lists';
import 'tinymce/plugins/preview';
import 'tinymce/plugins/searchreplace';
import 'tinymce/plugins/table';
import 'tinymce/plugins/wordcount';

import { Editor } from '@tinymce/tinymce-react';
import { PageContainer } from '@ant-design/pro-components';
import { App, Card, Typography } from 'antd';
import React, { useRef, useState } from 'react';
import { uploadFile } from '@/services/light-admin/file';

type TinyEditor = {
  getContent: () => string;
  setContent: (s: string) => void;
};

const INITIAL_HTML =
  '<h3>TinyMCE 7 示例</h3>' +
  '<p>工具栏里试试 <strong>加粗</strong>、<em>斜体</em>、列表、插入图片(走后端 <code>/files</code>)、全屏预览。</p>';

const TinymcePage: React.FC = () => {
  const { message } = App.useApp();
  const editorRef = useRef<TinyEditor | null>(null);
  const [html, setHtml] = useState(INITIAL_HTML);

  return (
    <PageContainer
      title="富文本编辑器(TinyMCE)"
      subTitle="TinyMCE 7 自托管 · GPL 授权 · 图片上传接 POST /files"
    >
      <Card style={{ marginBottom: 16 }}>
        <Editor
          licenseKey="gpl"
          onInit={(_, editor) => {
            editorRef.current = editor as unknown as TinyEditor;
          }}
          value={html}
          onEditorChange={setHtml}
          init={{
            height: 460,
            menubar: false,
            plugins:
              'advlist autolink lists link image code fullscreen preview searchreplace table wordcount',
            toolbar:
              'undo redo | blocks | bold italic underline strikethrough | ' +
              'alignleft aligncenter alignright | bullist numlist | ' +
              'link image table | searchreplace code preview fullscreen',
            branding: false,
            statusbar: true,
            skin: 'oxide',
            // We load skin JS statically above; skin CSS is loaded by TinyMCE itself.
            // Image upload hook — wires to our /files service.
            images_upload_handler: async (blobInfo) => {
              try {
                const res = await uploadFile(blobInfo.blob(), {
                  onProgress: () => {},
                });
                return res.url;
              } catch (err) {
                message.error('图片上传失败');
                throw err;
              }
            },
            automatic_uploads: true,
            paste_data_images: true,
            content_style:
              'body{font-family:AlibabaSans,Helvetica,Arial,sans-serif;font-size:14px;}',
          }}
        />
      </Card>

      <Card title="当前 HTML 输出">
        <Typography.Paragraph copyable>
          <code style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            {html}
          </code>
        </Typography.Paragraph>
      </Card>
    </PageContainer>
  );
};

export default TinymcePage;
