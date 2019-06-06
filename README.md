Программа для создания отчёта по графикам из графаны.

Generate doc report from Grafana dashboard.
![Generate](https://github.com/Maksimall89/generateDoc/blob/master/doc/gen_doc.jpg)

Вам нужно перед первым запуском сконфигурировать `tsconfig.json`. Вначале заполните поле "key" записав туда ваше [API Key](http://docs.grafana.org/http_api/auth/), затем заполните имя раздела "nameHeading", если оно пустое, то раздел не будет создан. Поле "namePanelsId" определяет название картинки, а поле "describePanel" описывает график. Параметр "panelsId" определяет какой график из дашборда Grafan нужно вставить "panelid" и ссылку на дашборд в поле "dashboward". Для этого просто скопируйте их из адресной строки браузера:  
![config_graf](https://github.com/Maksimall89/generateDoc/blob/master/doc/config_graf.jpg)  
Поле "panelsId" может содержать перечисление, тогда будут вставлены несколько графиков друг под другом. Поля "width" и "height" отвечают за ширину и длину графика, если они пустые или отсутствуют, тогда они будут игнорированы.

Все отчеты хранятся в папке `reports`. В папке `imgs` хранятся все скриншоты для последнего созданного отчёта. Папка `temlate` отвечает за шаблоны страниц генератора и отчёта.

Готовый exe файл из папки build следует запускать из корня проекта так как для его работы требуется наличие папок с шаблонами, конфигурацией запуска и т.п.

# Getting started
Get the source:

`go get github.com/maksimall89/generateDoc/...`

Compile:

`go build -o generateDoc.exe`

Start:

`execute generateDoc.exe`
`http://127.0.0.1:9005/`
