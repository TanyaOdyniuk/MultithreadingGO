Есть некоторое количество процессов, которые ставят задания (clients),
некоторое количество процессов, которые решают задачи (workers), 
и управляющий процесс, который в зависимости от количества и активности клиентов порождает или убивает обработчиков. 
Клиент называет имя случайного файла из тех, которые находяться в папке 20_newsgroup, а обработчик считает и возвращает клиенту количество слов в данном файле.

Реализовать эту систему.

Характеристики системы такие:

Система стартует с десятью клиентами и двумя обработчиками;

Клиент генерирует задачи с некоторой фиксированной частотой, которая может меняться (глобально), и которая в начале равна 50 ms, после этого он некоторое время (100 ms) ждет ответа. Если ответ не поступает, то он отменяет задачу;

Обработчик решает задачи по мере поступления (из некоторого общего пула). Для того, чтоб лучше видеть динамику, нужно ввести некоторую задержку после выполнения задачи -20 ms - для того, чтоб увеличение количества обработчиков повышало производительность системы, независимо от реального количества процессоров и распределения задач между ними;

Управляющий процесс регулирует количество обработчиков. Если количество отказов за последние 1000 ms превысило 20% от общего количества, то он создает нового обработчика. Если на протяжениии последних 1000 ms одновременно работало не болеее 50% от общего количества обработчиков, то он убивает одного из них;

Каждые 5 s система выводит в консоль количество работающих клиентов, обработчиков, уровень активности клиентов, а также количество сгенерированных за последние 5 s задач клиентами, а также количество решенных задач и отказов (абсолютно и в процентах);

Количество и активность клиентов (но не обработчиков) может регулироваться пользователем динамически: + добавляет клиента, - убирает, < уменьшает активность клиентов на 10%, > увеличивает (по нажатию этих клавиш делать подтверждающие записи в консоль, например +1 клиент (теперь 11)). Esc завершает выполнение;

Вынести значения фигурирующие в описании (начальное число клиентов и обработчиков, параметры определяющие активность, пороговые значения для управляющего процесса, ...) в константы в начале программмы;

Использовать только базовые примитивы для синхронизации go, chan, select, также для регулярных событий могут быть полезны каналы из модуля time.
