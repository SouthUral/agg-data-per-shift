# agg-data-per-shift

## Описание
Сервис агрегирует данные работы техники и записывает полученную информацию в БД.

## ToDo
 - [x] Подключение к RabbitMQ стриму;
 - [x] Чтение данных из стрима;
 - [x] Чтение env;
 - [ ] Обработка ошибок подключения, переподключение;

### Компоненты для запуска

Переменные окружения
```.env

# Postgres
ASD_POSTGRES_HOST="localhost"
ASD_POSTGRES_PORT="5435"
ASD_POSTGRES_DBNAME="report_bd"
SERVICE_PG_ILOGIC_USERNAME=<secret>
SERVICE_PG_ILOGIC_PASSWORD=<secret>

SERVICE_PG_NUM_PULL="10"


# RabbitMQ
ASD_RMQ_HOST="192.168.0.1"
ASD_RMQ_PORT="5432"
ASD_RMQ_VHOST="asd.asd.local.asd-test-03"
SERVICE_RMQ_ENOTIFY_USERNAME=<secret>
SERVICE_RMQ_ENOTIFY_PASSWORD=<secret>
SERVICE_RMQ_QUEUE="iLogic.Messages"
SERVICE_RMQ_NAME_CONSUMER="test_consumer"
```
conn=1/1    {minTime:5309 maxTime:10000676 sumTime:609746694 averageTime:121924 counterData:5001}
            {minTime:5484 maxTime:10000520 sumTime:589620722 averageTime:117829 counterData:5004}
            {minTime:5515 maxTime:10001002 sumTime:609090590 averageTime:121696 counterData:5005}
            {minTime:5483 maxTime:10000857 sumTime:653722375 averageTime:130171 counterData:5022}
            {minTime:5323 maxTime:10000981 sumTime:648740364 averageTime:129696 counterData:5002}

conn=1/10   {minTime:6115 maxTime:10001127 sumTime:334172063 averageTime:66607 counterData:5017}
            {minTime:5686 maxTime:10000984 sumTime:213433600 averageTime:42669 counterData:5002}
            {minTime:6972 maxTime:10001162 sumTime:201832907 averageTime:40237 counterData:5016}
            {minTime:6272 maxTime:10000981 sumTime:177259497 averageTime:35423 counterData:5004}
            {minTime:2642 maxTime:10001174 sumTime:103938945 averageTime:20700 counterData:5021}

conn=1/20   {minTime:7134 maxTime:10001563 sumTime:146935816 averageTime:29334 counterData:5009}
            {minTime:6274 maxTime:10000776 sumTime:120505489 averageTime:24052 counterData:5010}
            {minTime:5461 maxTime:10000782 sumTime:79820046 averageTime:15938 counterData:5008}
            {minTime:6461 maxTime:10001074 sumTime:105504201 averageTime:21083 counterData:5004}
            {minTime:6920 maxTime:10001126 sumTime:136284770 averageTime:27186 counterData:5013}

conn=5/10   {minTime:7719 maxTime:10019237 sumTime:350404237 averageTime:69469 counterData:5044}
            {minTime:6663 maxTime:10001106 sumTime:235343512 averageTime:46881 counterData:5020}
            {minTime:5664 maxTime:10001044 sumTime:166930171 averageTime:33352 counterData:5005}
            {minTime:7386 maxTime:10000845 sumTime:185732433 averageTime:37079 counterData:5009}
            {minTime:7393 maxTime:10000478 sumTime:180815154 averageTime:36134 counterData:5004}

conn=4/15   {minTime:5401 maxTime:10001029 sumTime:182255348 averageTime:36356 counterData:5013}
            {minTime:6687 maxTime:10001026 sumTime:165030711 averageTime:32979 counterData:5004}
            {minTime:6914 maxTime:10000323 sumTime:158108794 averageTime:31590 counterData:5005}
            {minTime:6928 maxTime:10000446 sumTime:139239468 averageTime:27814 counterData:5006}
            {minTime:5686 maxTime:10000609 sumTime:139935753 averageTime:27953 counterData:5006}

conn=3/6    {minTime:7289 maxTime:10001165 sumTime:272671335 averageTime:54414 counterData:5011}
            {minTime:5720 maxTime:10008724 sumTime:287487425 averageTime:57302 counterData:5017}
            
conn=10/20  {minTime:5458 maxTime:10000793 sumTime:76587328 averageTime:15296 counterData:5007}
            {minTime:5365 maxTime:10001187 sumTime:85294271 averageTime:17014 counterData:5013}
            {minTime:5283 maxTime:10000327 sumTime:70570986 averageTime:14077 counterData:5013}
            {minTime:5313 maxTime:10001120 sumTime:67604945 averageTime:13515 counterData:5002}
            {minTime:5394 maxTime:10001275 sumTime:132500300 averageTime:26347 counterData:5029}
            {minTime:4945 maxTime:10000415 sumTime:85580473 averageTime:17068 counterData:5014}

conn=20/30  {minTime:6227 maxTime:10000565 sumTime:119798862 averageTime:23921 counterData:5008}
            {minTime:5640 maxTime:10001110 sumTime:118013948 averageTime:23565 counterData:5008}
            {minTime:6739 maxTime:10000717 sumTime:119942375 averageTime:23954 counterData:5007}
            {minTime:6420 maxTime:10001275 sumTime:119951180 averageTime:23937 counterData:5011}
            {minTime:7447 maxTime:10000959 sumTime:119195527 averageTime:23810 counterData:5006}

conn=1/50   {minTime:5938 maxTime:10001016 sumTime:121614869 averageTime:24303 counterData:5004}
            {minTime:7504 maxTime:10001024 sumTime:215585981 averageTime:42902 counterData:5025}
            {minTime:6406 maxTime:10001101 sumTime:123921204 averageTime:24749 counterData:5007}
            {minTime:6530 maxTime:10000990 sumTime:118985578 averageTime:23759 counterData:5008}
            {minTime:6981 maxTime:10000450 sumTime:127513306 averageTime:25461 counterData:5008}

conn=10/50  {minTime:7021 maxTime:10000763 sumTime:122191507 averageTime:24365 counterData:5015}
            {minTime:7221 maxTime:10000402 sumTime:121850283 averageTime:24302 counterData:5014}
            {minTime:7060 maxTime:180217 sumTime:107452134 averageTime:21460 counterData:5007}
            {minTime:6171 maxTime:10001108 sumTime:130815588 averageTime:26053 counterData:5021}
            {minTime:5976 maxTime:10001102 sumTime:120705541 averageTime:24049 counterData:5019}
            
conn=20/50  {minTime:7489 maxTime:10000365 sumTime:108813474 averageTime:21710 counterData:5012}
            {minTime:7503 maxTime:10000796 sumTime:111400884 averageTime:22231 counterData:5011}
            {minTime:5829 maxTime:10001076 sumTime:127275399 averageTime:25389 counterData:5013}
            {minTime:6984 maxTime:10000783 sumTime:134848858 averageTime:26948 counterData:5004}
            {minTime:6640 maxTime:10001120 sumTime:140052382 averageTime:27932 counterData:5014}
            {minTime:5802 maxTime:10001096 sumTime:156970792 averageTime:31369 counterData:5004}

conn=30/40  {minTime:7132 maxTime:10004073 sumTime:114854576 averageTime:22920 counterData:5011}
            {minTime:7111 maxTime:10001418 sumTime:261521482 averageTime:52137 counterData:5016}
            {minTime:7324 maxTime:10001226 sumTime:129224195 averageTime:25793 counterData:5010}
            {minTime:6712 maxTime:10000672 sumTime:113997462 averageTime:22763 counterData:5008}
            {minTime:8005 maxTime:10001371 sumTime:140037842 averageTime:27951 counterData:5010}

conn=30/50  {minTime:7306 maxTime:10000667 sumTime:140186937 averageTime:27942 counterData:5017}
            {minTime:6732 maxTime:10002536 sumTime:144600169 averageTime:28885 counterData:5006}
            {minTime:7219 maxTime:10000743 sumTime:127928626 averageTime:25529 counterData:5011}
            {minTime:6828 maxTime:10001199 sumTime:132119320 averageTime:26350 counterData:5014}
            {minTime:6659 maxTime:10001241 sumTime:155890997 averageTime:31072 counterData:5017}

conn=30/60  {minTime:6924 maxTime:10000454 sumTime:121250913 averageTime:24221 counterData:5006}
            {minTime:7239 maxTime:10000842 sumTime:147775080 averageTime:29478 counterData:5013}
            {minTime:6817 maxTime:10001543 sumTime:127335492 averageTime:25456 counterData:5002}
            {minTime:6230 maxTime:10000363 sumTime:126437146 averageTime:25257 counterData:5006}
            {minTime:7706 maxTime:10000706 sumTime:130082524 averageTime:25985 counterData:5006}

conn=1/60   {minTime:7563 maxTime:10000363 sumTime:139553699 averageTime:27899 counterData:5002}