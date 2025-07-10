package fsadapter

// const defaultTemplate = `<html>
// <head>
// 	<title>{{.Title}}</title>
// </head>
// <body>
// 	<h1>{{.Title}}</h1>
// 	{{.PageContent}}

// 	<ul>
// 	{{range .Files}}
// 		<li><a href="http://example.com/file/{{.ID}}">{{.Name}}</a></li>
// 	{{end}}
// 	</ul>
// </body>
// </html>
// `

const defaultTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.3.0/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-icons/1.10.0/font/bootstrap-icons.min.css" rel="stylesheet">
    <style>
        .file-list {
            max-height: 600px;
            overflow-y: auto;
        }
        .file-item {
            border-bottom: 1px solid #e9ecef;
            padding: 0.75rem 0;
        }
        .file-item:last-child {
            border-bottom: none;
        }
        .torrent-info {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-radius: 10px;
            padding: 2rem;
            margin-bottom: 2rem;
        }
        .description-card {
            background: #f8f9fa;
            border-left: 4px solid #007bff;
        }
    </style>
</head>
<body class="bg-light">
    <div class="container my-4">
        <!-- Заголовок раздачи -->
        <div class="torrent-info">
            <h1 class="mb-3">
                <i class="bi bi-cloud-download me-2"></i>
                {{.Title}}
            </h1>
            <div class="text-center">
                <span class="badge bg-success fs-6">
                    <i class="bi bi-files me-1"></i>
                    {{len .Files}} файлов
                </span>
            </div>
        </div>

        <!-- Описание раздачи -->
        {{if .PageContent}}
        <div class="card description-card mb-4">
            <div class="card-body">
                <h5 class="card-title">
                    <i class="bi bi-info-circle me-2"></i>
                    Описание
                </h5>
                <div class="card-text">
                    {{.PageContent}}
                </div>
            </div>
        </div>
        {{end}}

        <!-- Список файлов -->
        <div class="card">
            <div class="card-header bg-primary text-white">
                <h5 class="mb-0">
                    <i class="bi bi-folder-open me-2"></i>
                    Файлы для скачивания
                </h5>
            </div>
            <div class="card-body p-0">
                {{if .Files}}
                <div class="file-list">
                    {{range $index, $file := .Files}}
                    <div class="file-item px-3" data-file-id="{{$file.ID}}">
                        <div class="d-flex align-items-center justify-content-between">
                            <div class="d-flex align-items-center flex-grow-1">
                                <i class="bi bi-file-earmark text-muted me-2"></i>
                                <span class="file-name">{{$file.Name}}</span>
                            </div>
                            <div class="d-flex align-items-center ms-3">
                                <div class="me-3">
                                    <small class="text-muted">Скачиваний:</small>
                                    <span class="badge bg-info download-count" data-file-id="{{$file.ID}}">
                                        <i class="bi bi-hourglass-split"></i>
                                    </span>
                                </div>
                                <a href="http://{{$.URL}}/file/{{$file.ID}}" 
                                   class="btn btn-outline-primary btn-sm">
                                    <i class="bi bi-download me-1"></i>
                                    Скачать
                                </a>
                            </div>
                        </div>
                    </div>
                    {{end}}
                </div>
                {{else}}
                <div class="text-center py-4">
                    <i class="bi bi-folder-x text-muted" style="font-size: 3rem;"></i>
                    <p class="text-muted mt-2">Файлы не найдены</p>
                </div>
                {{end}}
            </div>
        </div>

    </div>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.3.0/js/bootstrap.bundle.min.js"></script>
    <script>
        // Загрузка счетчиков скачиваний
        async function loadDownloadCounts() {
            try {
                const response = await fetch('http://{{.URL}}/info/{{.ID}}');
                const data = await response.json();
                
                // Обновляем счетчики для каждого файла
                document.querySelectorAll('.download-count').forEach(element => {
                    const fileId = element.getAttribute('data-file-id');
                    const count = data[fileId] || 0;
                    element.innerHTML = count;
                    element.className = count > 0 ? 'badge bg-success' : 'badge bg-secondary';
                });
            } catch (error) {
                console.error('Ошибка загрузки счетчиков:', error);
                // В случае ошибки показываем 0
                document.querySelectorAll('.download-count').forEach(element => {
                    element.innerHTML = '0';
                    element.className = 'badge bg-secondary';
                });
            }
        }

        // Загружаем счетчики при загрузке страницы
        document.addEventListener('DOMContentLoaded', loadDownloadCounts);
    </script>
</body>
</html>
`
