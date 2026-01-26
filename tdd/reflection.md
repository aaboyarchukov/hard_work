# TDD

## Вводное

**TDD** (Test-Driven Development) - подход в разработке программного обеспечения, который ведется через тестирование. 

**Модульное тестирование** (-юнит тестирование) - подход  при котором тестируется минимальная единица разработанной части проекта, которая является изолируемой.

**Регрессионное тестирование** - подход, с помощью которого проверяется верная, с точки зрения спецификации, функциональность всей системы при ее изменении.

**Интеграционное тестирование** - подход, с помощью которого тестируются модули системы между собой, то есть во взаимодействии.

Суть TDD подхода заключается в разработке строгой спецификации для написания частей программного обеспечения. Нужно это для того, что писать правильный, чистый, минималистичный и легко масштабируемый код.

Я понял, что при разрабокте по данному методу ты проходишь три стадии до тех пор пока твой код не будет покрывать все тест-кейсы:

1. **Red** - первый этап, в котором необходимо в тестах отразить кейсы, которые должен обработать твой будущий код. После написания тебе необходимо их запустить, удостоверевшись в том, что они собираются, и увидеть, что все они падают, показывая, что твой код не соответствует заданной тобой спецификации.
2. **Green** - после первой стадии, необходимо писать, собственно, сам функционал, и прогонять его через написанную спецификацию раннее. При обнаружении ошибок и неточностей - идут исправления, и так до тех пор, пока код не станет соответствовать спецификации. Далее, когда все ошибки исправлены - идет завершающая стадия - рефактор.
3. **Refactor** - стадия на которой мы приводим свой код в порядок, улучшаем его и оптимизируем, при том, что он все также должен соответсвовать спецификации.

Но также важн формулировать эти кейсы, вот основные из них, которые встречаются и их необходимо обработать практически везде:
- happy path
- границы (0, 1, max, пусто)
- ошибки (invalid input)
- идемпотентность/повторный вызов
- порядок операций (если важен)

То есть при написании ПО по TDD - мы определяем спецификацию кода, через описание тестов. Но, как и везде, важно быть прогматиком и работать по данной методике только там, где необходимо, а также стоит строго следовать ей. Ведь, например, если мы будем писать функционал, до определения спецификации, через тесты - мы нарушаем основопологающие принципы, ведь не логично, сначала писать функционал, не зная какой спецификации он должен следовать.

В общем, все надо применять с умом и по необходимости.

## Преимущества

1. Код становится чище и более оптимизирован, так как проходит несколько стадий проверки - улучшение дизайна кода
2. Написанный функционал, проходит проверки, перед тем, как использоваться дальше, что гарантирует его исправность в большинстве случаев

## Недостатки

1. Уходит больше времени на разработку (хотя оно модет компенсировать в дальнейшем)

## Что необходимо изучить:

1. Интеграционные тесты
2. Тестирование **экспортируемого API пакета** (black-box) и тестирование внутренних деталей через (white-box)
3. Моки/стабы/фейки DI в Go
4. Фикстуры (fixture)

## Примеры написания кода по TDD:

Сейчас я придерживаюсь написания кода по TDD при изучении алгоритмов и структур данных:

```python
import unittest

from task1 import SelectionSortStep, BubbleSortStep


class TestSelectionSortStep(unittest.TestCase):

    def test_one_step_sorting(self):
        cases = [
            ([3, 1, 2], 0, [1, 3, 2]),
            ([3, 1, 2], 1, [3, 1, 2]),
            ([5, 4, 3, 2, 1], 0, [1, 4, 3, 2, 5]),
            ([2, 2, 1], 0, [1, 2, 2]),
            ([1, 2, 3], 0, [1, 2, 3]),
        ]

        for arr, i, expected in cases:
            with self.subTest(arr=arr, i=i):
                arr_copy = arr.copy()
                SelectionSortStep(arr_copy, i)

                self.assertEqual(
                    arr_copy, expected,
                    msg=(
                        f"FAIL: Один шаг сортировки (i={i}) выполнен неверно.\n"
                        f"Вход:      {arr}\n"
                        f"Ожидалось: {expected}\n"
                        f"Получено:  {arr_copy}"
                    )
                )

    def test_n_steps_sorting(self):
        cases = [
            ([3, 1, 2, 5, 4], 2),
            ([5, 4, 3, 2, 1], 3),
            ([2, 2, 1, 3, 1], 4),
            ([1, 2, 3, 4], 1),
        ]

        for arr, k in cases:
            with self.subTest(arr=arr, k=k):
                arr_copy = arr.copy()
                expected_prefix = sorted(arr)[:k] # trusted

                for i in range(k):
                    SelectionSortStep(arr_copy, i)

                self.assertEqual(
                    arr_copy[:k], expected_prefix,
                    msg=(
                        f"FAIL: После {k} шагов первые {k} элементов неверны.\n"
                        f"Вход:                 {arr}\n"
                        f"Ожидался префикс:     {expected_prefix}\n"
                        f"Полученный префикс:   {arr_copy[:k]}\n"
                        f"Текущее состояние arr_copy:  {arr_copy}"
                    )
                )
                
                self.assertCountEqual(
                    arr_copy, arr,
                    msg=(
                        f"FAIL: После {k} шагов изменилось содержимое массива (потеря/дублирование элементов).\n"
                        f"Вход:     {arr}\n"
                        f"Получено: {arr_copy}"
                    )
                )

    def test_full_array_sorting(self):
        cases = [
            [],
            [1],
            [2, 1],
            [3, 1, 2],
            [5, 4, 3, 2, 1],
            [2, 2, 1, 3, 1],
            [1, 2, 3, 4, 5],
        ]

        for arr in cases:
            with self.subTest(arr=arr):
                arr_copy = arr.copy()
                for i in range(0, len(arr_copy)):
                    SelectionSortStep(arr_copy, i)

                self.assertEqual(
                    arr_copy, sorted(arr),
                    msg=(
                        "FAIL: Полная сортировка массива шагами selection sort дала неверный результат.\n"
                        f"Вход:      {arr}\n"
                        f"Ожидалось: {sorted(arr)}\n"
                        f"Получено:  {arr_copy}"
                    )
                )

class TestBubbleSortStep(unittest.TestCase):

    def test_one_step_sorting(self):
        cases = [
            # (input_array, expected_array_after_one_pass, expected_return)
            ([3, 2, 1], [2, 1, 3], False),
            ([1, 3, 2], [1, 2, 3], False),
            ([1, 2, 3], [1, 2, 3], True),
            ([2, 2, 1], [2, 1, 2], False),
            ([1], [1], True),
            ([], [], True),
            ([5, 1, 4, 2, 8], [1, 4, 2, 5, 8], False),
        ]

        for arr, expected_arr, expected_flag in cases:
            with self.subTest(arr=arr):
                arr_copy = arr.copy()
                flag = BubbleSortStep(arr_copy)

                self.assertEqual(
                    arr_copy, expected_arr,
                    msg=(
                        "FAIL: Один проход BubbleSortStep дал неверный массив.\n"
                        f"Вход:      {arr}\n"
                        f"Ожидалось: {expected_arr}\n"
                        f"Получено:  {arr_copy}"
                    )
                )
                self.assertEqual(
                    flag, expected_flag,
                    msg=(
                        "FAIL: BubbleSortStep вернул неверный флаг (True если не было обменов).\n"
                        f"Вход:           {arr}\n"
                        f"Ожидался флаг:   {expected_flag}\n"
                        f"Полученный флаг: {flag}\n"
                        f"Состояние a:     {arr_copy}"
                    )
                )

    def test_n_steps_sorting(self):
        cases = [
            ([3, 1, 2, 5, 4], 2),
            ([5, 4, 3, 2, 1], 3),
            ([2, 2, 1, 3, 1], 4),
            ([1, 2, 3, 4], 1),
        ]

        for arr, k in cases:
            with self.subTest(arr=arr, k=k):
                arr_copy = arr.copy()
                sorted_arr = sorted(arr)

                last_flag = None
                for _ in range(k):
                    last_flag = BubbleSortStep(arr_copy)

                self.assertEqual(
                    arr_copy[-k:] if k > 0 else [],
                    sorted_arr[-k:] if k > 0 else [],
                    msg=(
                        f"FAIL: После {k} проходов последние {k} элементов неверны.\n"
                        f"Вход:                 {arr}\n"
                        f"Ожидался суффикс:     {sorted_arr[-k:] if k > 0 else []}\n"
                        f"Полученный суффикс:   {arr_copy[-k:] if k > 0 else []}\n"
                        f"Текущее состояние a:  {arr_copy}\n"
                        f"Последний флаг:       {last_flag}"
                    )
                )

                self.assertCountEqual(
                    arr_copy, arr,
                    msg=(
                        f"FAIL: После {k} проходов изменилось содержимое массива (потеря/дублирование элементов).\n"
                        f"Вход:     {arr}\n"
                        f"Получено: {arr_copy}"
                    )
                )

    def test_full_array_sorting(self):
        cases = [
            [],
            [1],
            [2, 1],
            [3, 1, 2],
            [5, 4, 3, 2, 1],
            [2, 2, 1, 3, 1],
            [1, 2, 3, 4, 5],
        ]

        for arr in cases:
            with self.subTest(arr=arr):
                arr_copy = arr.copy()

                max_passes = max(1, len(arr_copy))
                sorted_expected = sorted(arr)

                finished = False
                for pass_no in range(max_passes):
                    no_swaps = BubbleSortStep(arr_copy)
                    if no_swaps:
                        finished = True
                        break

                self.assertTrue(
                    finished,
                    msg=(
                        "FAIL: BubbleSortStep не сообщил о завершении (True) за ожидаемое число проходов.\n"
                        f"Вход:                 {arr}\n"
                        f"Ожидаемая сортировка: {sorted_expected}\n"
                        f"Текущее состояние a:  {arr_copy}\n"
                        f"Лимит проходов:       {max_passes}"
                    )
                )

                self.assertEqual(
                    arr_copy, sorted_expected,
                    msg=(
                        "FAIL: Полная сортировка пузырьком (повторяя BubbleSortStep) дала неверный результат.\n"
                        f"Вход:      {arr}\n"
                        f"Ожидалось: {sorted_expected}\n"
                        f"Получено:  {arr_copy}"
                    )
                )

if __name__ == "__main__":
    unittest.main()
```

```python
# 1. get func of get min of two element
def index_of_min(a, indx_a, b, indx_b):
    if a < b:
        return indx_a
    return indx_b

# 2. get func of get index of min element
def indx_of_min_in_array(array, start_indx):
    result = start_indx
    while start_indx < len(array) - 1:
        result = index_of_min(array[result], result, array[start_indx + 1], start_indx + 1)
        start_indx += 1
    return result

# 3. get func of selection sort step
# t = O(n), where n = len(array) 
# mem = O(1)
def SelectionSortStep(array : list, i : int):
    indx_min = indx_of_min_in_array(array, i)
    array[i], array[indx_min] = array[indx_min], array[i]

# 4. get func of bubble sort step
# t = O(n), where n = len(array) 
# mem = O(1)
def BubbleSortStep(array):
    default_switches = 0

    amount_switches = default_switches
    for indx in range(0, len(array) - 1):
        if array[indx] > array[indx + 1]:
            array[indx], array[indx + 1] = array[indx + 1], array[indx]
            amount_switches += 1
    
    return amount_switches == default_switches
```

```python
import unittest

from task2 import InsertionSortStep


class TestInsertionSortStep(unittest.TestCase):

    def test_one_step_sorts_chain_fully(self):
        cases = [
            ([3, 1, 2], 1, 1, [3, 1, 2]),
            ([3, 2, 1], 1, 1, [3, 1, 2]),
            ([7, 6, 5, 4, 3, 2, 1], 3, 0, [1, 6, 5, 4, 3, 2, 7]),
            ([9, 8, 7, 6, 5, 4, 3, 2], 3, 1, [9, 2, 7, 6, 5, 4, 3, 8]),
            ([2, 1, 1, 1], 1, 1, [2, 1, 1, 1]),
        ]

        for arr, step, i, expected in cases:
            with self.subTest(arr=arr, step=step, i=i):
                a = arr.copy()
                InsertionSortStep(a, step, i)
                self.assertEqual(
                    a, expected,
                    msg=(
                        "FAIL: Один шаг (полная сортировка подпоследовательности i, i+step, ...) выполнен неверно.\n"
                        f"Вход:      {arr}\n"
                        f"step={step}, i={i}\n"
                        f"Ожидалось: {expected}\n"
                        f"Получено:  {a}"
                    )
                )

    def test_one_step_does_not_touch_elements_outside_chain(self):
        arr = [7, 6, 5, 4, 3, 2, 1]
        step = 3
        i = 0  # цепочка 0,3,6

        a = arr.copy()
        InsertionSortStep(a, step, i)

        chain = set(range(i, len(arr), step))
        for idx in range(len(arr)):
            if idx not in chain:
                self.assertEqual(
                    a[idx], arr[idx],
                    msg=(
                        "FAIL: Изменён элемент вне сортируемой подпоследовательности.\n"
                        f"Вход: {arr}\n"
                        f"step={step}, i={i}\n"
                        f"Цепочка индексов: {sorted(chain)}\n"
                        f"Индекс {idx}: было {arr[idx]} -> стало {a[idx]}\n"
                        f"Итоговый массив: {a}"
                    )
                )

        self.assertCountEqual(
            a, arr,
            msg=(
                "FAIL: После шага изменилось содержимое массива (потеря/дублирование элементов).\n"
                f"Вход:     {arr}\n"
                f"Получено: {a}"
            )
        )

    def test_n_steps_for_fixed_step_run_i_0_to_step_minus_1(self):
        cases = [
            ([7, 6, 5, 4, 3, 2, 1], 3),
            ([10, 9, 8, 7, 6, 5, 4, 3, 2, 1], 4),
            ([2, 2, 1, 3, 1, 0, 0], 2),
            ([1, 5, 3, 5, 2, 5, 4], 3),
        ]

        for arr, step in cases:
            with self.subTest(arr=arr, step=step):
                a = arr.copy()

                for i in range(min(step, len(a))):
                    InsertionSortStep(a, step, i)

                for r in range(min(step, len(a))):
                    seq = [a[j] for j in range(r, len(a), step)]
                    self.assertEqual(
                        seq, sorted(seq),
                        msg=(
                            "FAIL: После шагов i=0..step-1 подпоследовательность по r не отсортирована.\n"
                            f"Исходный вход: {arr}\n"
                            f"step={step}, r={r}\n"
                            f"Подпоследовательность: {seq}\n"
                            f"Ожидалось:            {sorted(seq)}\n"
                            f"Итоговый массив:      {a}"
                        )
                    )

                self.assertCountEqual(
                    a, arr,
                    msg=(
                        "FAIL: Изменилось содержимое массива (потеря/дублирование элементов) после n шагов.\n"
                        f"Вход:     {arr}\n"
                        f"Получено: {a}"
                    )
                )

    def test_n_steps_for_fixed_step_run_i_0_to_n_minus_1(self):
        cases = [
            ([7, 6, 5, 4, 3, 2, 1], 3),
            ([10, 9, 8, 7, 6, 5, 4, 3, 2, 1], 4),
            ([2, 2, 1, 3, 1, 0, 0], 2),
            ([1, 5, 3, 5, 2, 5, 4], 3),
        ]

        for arr, step in cases:
            with self.subTest(arr=arr, step=step):
                a = arr.copy()

                for i in range(len(a)):
                    InsertionSortStep(a, step, i)

                for r in range(min(step, len(a))):
                    seq = [a[j] for j in range(r, len(a), step)]
                    self.assertEqual(
                        seq, sorted(seq),
                        msg=(
                            "FAIL: После прогона i=0..n-1 подпоследовательность по r не отсортирована.\n"
                            f"Исходный вход: {arr}\n"
                            f"step={step}, r={r}\n"
                            f"Подпоследовательность: {seq}\n"
                            f"Ожидалось:            {sorted(seq)}\n"
                            f"Итоговый массив:      {a}"
                        )
                    )

                self.assertCountEqual(
                    a, arr,
                    msg=(
                        "FAIL: Изменилось содержимое массива (потеря/дублирование элементов) после n шагов.\n"
                        f"Вход:     {arr}\n"
                        f"Получено: {a}"
                    )
                )

    def test_full_array_sorting_shell_like_schedule(self):
        cases = [
            [],
            [1],
            [2, 1],
            [7, 6, 5, 4, 3, 2, 1],
            [5, 1, 4, 2, 8],
            [2, 2, 1, 3, 1, 0, 0],
            [1, 2, 3, 4, 5],
        ]

        for arr in cases:
            with self.subTest(arr=arr):
                a = arr.copy()

                step = len(a) // 2
                while step > 0:
                    for i in range(min(step, len(a))):
                        InsertionSortStep(a, step, i)
                    step //= 2

                self.assertEqual(
                    a, sorted(arr),
                    msg=(
                        "FAIL: Полная сортировка (Shell-подобной схемой через InsertionSortStep) дала неверный результат.\n"
                        f"Вход:      {arr}\n"
                        f"Ожидалось: {sorted(arr)}\n"
                        f"Получено:  {a}"
                    )
                )


if __name__ == "__main__":
    unittest.main(verbosity=2)
```

```python
# t = O(n // step), where n = len(array) 
# mem = O(1)
def find_position(array, start, stop, step):
    init_pos = start
    while start > stop and array[start - step] > array[init_pos]:
        start -= step

    return start

# t = O(n // step), where n = len(array) 
# mem = O(1)
def insert_at_position(array, position_from, position_to, step):
    target_value = array[position_from]
    target_position = position_to

    while position_to < position_from:
        array[position_from] = array[position_from - step]
        position_from -= step
    
    array[target_position] = target_value
    print(array)

# t = O(((n - i) // step) ^ 2), where n = len(array) 
# mem = O(1)
def InsertionSortStep(array, step, i):
    for indx in range(i + step, len(array), step):
        target_position = find_position(array, indx, i, step)
        insert_at_position(array, indx, target_position, step)
```

## После прочтения материалов:

Главные принципы, которым необходимо следовать при проектировании систем:
1. Не смешивать уровни рассуждений о системе
2. Начинать проектировать с логического уровня
3. Следовать и отталкиваться от одной спецификации (дизайн/архитектура)
4. Ни тесты должны следовать коду, ни код тестам - они должны следовать одному дизайну

Для того, чтобы переписать наш код, определим спецификацию:

Нам необходимо разработать функцию, которая выполняет один шаг для сортировки вставками. 

Для реализации этой функции нам понадобятся дополнительные функции:

### 1. Поиск целевого места

  `find_position(array, start, stop, step)`

**Назначение:** Найти правильную позицию для вставки элемента в отсортированной подпоследовательности.

**Параметры:**  
- `array: List[int]` - входной массив  
- `start: int` - начальная позиция поиска (позиция вставляемого элемента)  
- `stop: int` - конечная позиция поиска (граница отсортированной части)  
- `step: int` - шаг между сравниваемыми элементами

**Возвращает:**  
- `int` - позицию, куда следует вставить элемент

**Предусловия:**  
- `0 <= stop <= start < len(array)`  
- `step > 0`  

**Постусловия:**  
- Возвращаемая позиция находится в диапазоне `[stop, start]`  

### 2. Функция сдвига элементов

`insert_at_position(array, position_from, position_to, step)`

**Назначение:** Переместить элемент из одной позиции в другую, сдвигая промежуточные элементы.

**Параметры:**  
- `array: List[int]` - массив для модификации  
- `position_from: int` - исходная позиция элемента  
- `position_to: int` - целевая позиция для вставки  
- `step: int` - шаг между элементами

**Возвращает:** `None` (изменяет массив in-place)

**Предусловия:**  
- `0 <= position_to <= position_from < len(array)`  
- `step > 0` 

**Постусловия:**  
- Элемент перемещен с `position_from` на `position_to`  
- Элементы между позициями сдвинуты на `step` позиций вправо  
- Порядок остальных элементов не изменен

### 3. Целевая функция для шага сортировки

 `InsertionSortStep(array, step, i)`

**Назначение:** Выполнить один шаг сортировки вставками для подпоследовательности с заданным шагом.

**Параметры:**  
- `array: List[int]` - массив для сортировки  
- `step: int` - шаг между элементами подпоследовательности  
- `i: int` - начальная позиция подпоследовательности

**Возвращает:** `None` (изменяет массив in-place)

**Предусловия:**  
- `0 <= i < len(array)`  
- `step > 0`  
- `i + step <= len(array)` (есть хотя бы один элемент для сортировки)

**Постусловия:**  
- Подпоследовательность `array[i], array[i+step], array[i+2*step], ...`отсортирована по возрастанию  
- Элементы других подпоследовательностей не изменены

Теперь напишем код по TDD из предыдущего пункта именно с учетом спецификации!
И соответственно главный вывод в том, что тесты и реализация должны меняться только тогда, когда меняется сама **спецификация**, а не когда меняется один из компонентов (из реализации и тестов). Итоговый результат:

```python
import unittest

from task2 import InsertionSortStep, find_position, insert_at_position

class TestFindPosition(unittest.TestCase):
    # preconditionals
    def test_find_position_preconditionals(self):
        cases = [
            ([1, 2, 3], [2, 0, 1]),
            ([1, 2, 3, 4], [3, 1, 2]),
            ([5], [0, 0, 1]),
        ]

        for case in cases:
            with self.subTest(case=case):
                array, (start, stop, step) = case
                self.assertGreaterEqual(stop, 0, "stop должен быть >= 0")
                self.assertLessEqual(stop, start, "stop должен быть <= start") 
                self.assertLess(start, len(array), "start должен быть < len(array)")
                self.assertGreater(step, 0, "step должен быть > 0")

    # postconditionals
    def test_find_position_postconditions(self):
        cases = [
            ([1, 2, 3], [2, 0, 1]),
            ([5, 3, 1], [2, 0, 1]), 
            ([1, 3, 5], [2, 0, 1]),
            ([2, 2, 2], [1, 0, 1]),
        ]

        for case in cases:
            with self.subTest(case=case):
                array, (start, stop, step) = case
                result = find_position(array, start, stop, step)

                self.assertGreaterEqual(result, stop, 
                    f"Результат {result} должен быть >= stop={stop}")
                self.assertLessEqual(result, start,
                    f"Результат {result} должен быть <= start={start}")

    # edge cases: find_position
    def test_find_position_edge_cases(self):
        cases = [
            ([5], [0, 0, 1], 0),  # Минимальный массив
            ([1, 2, 3], [2, 2, 1], 2),  # start == stop
            ([5, 3, 1], [2, 0, 1], 0),  # Элемент меньше всех
            ([1, 3, 5], [2, 0, 1], 2),  # Элемент больше всех
            ([1, 9, 2, 8, 3, 7], [4, 0, 2], 4),  # Шаг больше 1
        ]

        for case in cases:
            with self.subTest(case=case):
                array, (start, stop, step), expected = case

                result = find_position(array, start, stop, step)
                if isinstance(expected, list):
                    self.assertIn(result, expected)
                else:
                    self.assertEqual(result, expected)
    
class TestInsertAtPosition(unittest.TestCase):
    # preconditionals
    def test_insert_at_position_preconditions(self):
        cases = [
            ([1, 2, 3], [2, 0, 1]),
            ([5], [0, 0, 1]),
            ([1, 2, 3, 4, 5, 6], [4, 2, 2]),
        ]

        for case in cases:
            with self.subTest(case=case):
                array, (position_from, position_to, step) = case

                self.assertGreaterEqual(position_to, 0, "position_to должен быть >= 0")
                self.assertLessEqual(position_to, position_from, "position_to должен быть <= position_from")
                self.assertLess(position_from, len(array), "position_from должен быть < len(array)")
                self.assertGreater(step, 0, "step должен быть > 0")

    # postconditionals
    def test_insert_at_position_postconditions(self):
        cases = [
            ([1, 3, 2], [2, 1, 1]),
            ([2, 3, 4, 1], [3, 0, 1]),
            ([3, 9, 1, 8, 2, 7], [4, 0, 2]),
        ]

        for case in cases:
            with self.subTest(case=case):
                array, (position_from, position_to, step) = case

                original_value = array[position_from]
                original_content = array.copy()
                original_length = len(array)

                insert_at_position(array, position_from, position_to, step)

                # Проверяем постусловия
                self.assertEqual(array[position_to], original_value, 
                    "Элемент не перемещен на целевую позицию")
                self.assertEqual(len(array), original_length,
                    "Изменилось количество элементов")
                self.assertCountEqual(array, original_content,
                    "Изменилось содержимое массива")

    # edge cases: insert_at_position
    def test_insert_at_position_edge_cases(self):
        cases = [
            ([1, 2, 3], [1, 1, 1], [1, 2, 3]),  # Нет перемещения
            ([1, 3, 2], [2, 1, 1], [1, 2, 3]),  # Перемещение на одну позицию
            ([2, 1, 3], [1, 0, 1], [1, 2, 3]),  # Перемещение в начало
            ([2, 3, 4, 1], [3, 0, 1], [1, 2, 3, 4]),  # Максимальное смещение
            ([3, 9, 1, 8, 2, 7], [4, 0, 2], [2, 9, 3, 8, 1, 7]),  # Шаг больше 1
            ([1], [0, 0, 1], [1]),  # Минимальный массив
        ]

        for case in cases:
            with self.subTest(case=case):
                original_array, (position_from, position_to, step), expected = case
                array = original_array.copy()

                insert_at_position(array, position_from, position_to, step)
                self.assertEqual(array, expected)
    

class TestInsertionSortStep(unittest.TestCase):

    # preconditionals
    def test_insertion_sort_step_preconditions(self):
        cases = [
            ([1, 2, 3], [1, 0]),
            ([5], [1, 0]),
            ([1, 2, 3, 4], [2, 1]),
            ([1, 2, 3, 4, 5], [3, 2]),
        ]

        for case in cases:
            with self.subTest(case=case):
                array, (step, i) = case

                self.assertGreaterEqual(i, 0, "i должен быть >= 0")
                self.assertLess(i, len(array), "i должен быть < len(array)")
                self.assertGreater(step, 0, "step должен быть > 0")

    # postconditionals
    def test_insertion_sort_step_postconditions(self):
        cases = [
            ([3, 9, 1, 8, 2, 7], [2, 0]),
            ([4, 5, 2, 6, 1, 7], [3, 1]),
            ([5, 2, 4, 1, 3], [2, 0]),
        ]

        for case in cases:
            with self.subTest(case=case):
                array, (step, i) = case

                original_content = array.copy()
                original_length = len(array)

                # Запоминаем элементы других подпоследовательностей
                other_elements = {}
                for idx in range(len(array)):
                    if idx % step != i % step:
                        other_elements[idx] = array[idx]

                InsertionSortStep(array, step, i)

                # 1. Подпоследовательность отсортирована
                subsequence = [array[j] for j in range(i, len(array), step)]
                self.assertEqual(subsequence, sorted(subsequence),
                    "Подпоследовательность должна быть отсортирована")

                # 2. Элементы других подпоследовательностей не изменены
                for idx, expected_value in other_elements.items():
                    self.assertEqual(array[idx], expected_value,
                        f"Элемент вне подпоследовательности изменен: индекс {idx}")

                # 3. Содержимое массива сохранилось
                self.assertEqual(len(array), original_length,
                    "Изменилось количество элементов")
                self.assertCountEqual(array, original_content,
                    "Изменилось содержимое массива")

    # edge cases: InsertionSortStep
    def test_insertion_sort_step_edge_cases(self):
        """Краевые случаи для InsertionSortStep"""
        cases = [
            ([5], [1, 0], [5]),  # Минимальный массив
            ([1, 2], [1, 0], [1, 2]),  # Два элемента, отсортированы
            ([2, 1], [1, 0], [1, 2]),  # Два элемента, нужна сортировка
            ([3, 1, 2], [2, 1], [3, 1, 2]),  # Один элемент в подпоследовательности
            ([1, 2, 3], [3, 0], [1, 2, 3]),  # Шаг равен длине массива
            ([1, 3, 2, 4], [2, 1], [1, 3, 2, 4]),  # Уже отсортированная подпоследовательность
            ([5, 5, 5, 5], [1, 0], [5, 5, 5, 5]),  # Все элементы одинаковые
            ([4, 9, 3, 8, 2, 7, 1, 6], [2, 0], [1, 9, 2, 8, 3, 7, 4, 6]),  # Worst case
            ([1, 9, 2, 8, 3, 7, 4, 6], [2, 0], [1, 9, 2, 8, 3, 7, 4, 6]),  # Best case
            ([1, 2, 3, 4, 5], [3, 2], [1, 2, 3, 4, 5]),  # Граничный случай
        ]

        for case in cases:
            with self.subTest(case=case):
                original_array, (step, i), expected = case
                array = original_array.copy()

                InsertionSortStep(array, step, i)
                self.assertEqual(array, expected)

class TestFullArraySorting(unittest.TestCase):
    def test_full_array_sorting_edge_cases(self):
        """Краевые случаи для полной сортировки"""
        cases = [
            ([], [], "empty_array"),
            ([1], [1], "single_element"),
            ([2, 1], [1, 2], "two_elements_reversed"),
            ([1, 2], [1, 2], "two_elements_sorted"),
            ([7, 6, 5, 4, 3, 2, 1], [1, 2, 3, 4, 5, 6, 7], "reverse_order_worst_case"),
            ([1, 2, 3, 4, 5], [1, 2, 3, 4, 5], "already_sorted_best_case"),
            ([5, 1, 4, 2, 8], [1, 2, 4, 5, 8], "random_order"),
            ([2, 2, 1, 3, 1, 0, 0], [0, 0, 1, 1, 2, 2, 3], "duplicates_present"),
            ([5, 5, 5, 5], [5, 5, 5, 5], "all_elements_equal"),
            ([3], [3], "single_element_boundary"),
        ]

        for case in cases:
            with self.subTest(test=case[2], original=case[0]):
                original_array, expected, description = case
                array = original_array.copy()

                step = len(array) // 2
                while step > 0:
                    for i in range(min(step, len(array))):
                        if i < len(array):
                            InsertionSortStep(array, step, i)
                    step //= 2

                self.assertEqual(array, expected,
                    f"Сортировка {description} дала неверный результат")

if __name__ == "__main__":
    unittest.main(verbosity=2)
```

```python
# t = O(n // step), where n = len(array) 
# mem = O(1)
def find_position(array: list[int], start, stop, step):
    init_pos = start
    while start > stop and array[start - step] > array[init_pos]:
        start -= step

    return start

# t = O(n // step), where n = len(array) 
# mem = O(1)
def insert_at_position(array: list[int], position_from, position_to, step):
    target_value = array[position_from]
    target_position = position_to

    while position_to < position_from:
        array[position_from] = array[position_from - step]
        position_from -= step
    
    array[target_position] = target_value

# t = O(((n - i) // step) ^ 2), where n = len(array) 
# mem = O(1)
def InsertionSortStep(array: list[int], step, i):
    for indx in range(i + step, len(array), step):
        target_position = find_position(array, indx, i, step)
        insert_at_position(array, indx, target_position, step)
```

## Что изменилось

После введения спецификации и рефакторинга кода:

- тесты и реализация следуют строго спецификации
- исключил из тестов специфические случаи
- включил в тесты проверки -пост и -пред условий, согласно спецификации

## Выводы

После изучения материалов вывел для себя главную идею:

Что последовательность разработки четко определяется следующим образом:

**Спецификация → Тесты → Реализация**

1. **Спецификация** - единый источник истины, описывающий ЧТО должна делать программа
2. **Тесты** - проверяют соответствие спецификации 
3. **Реализация** - КАК выполнить требования спецификации

Но также важно понимать, что не везде нужно следовать TDD, то есть отталкиваться именно от требований к проекту. Но что точно должно быть везде: спецификация - ежиный источник истины, и все должно начинаться с нее.
Также важно начинать с логического уровня проектирования, и не смешивать другие уровни между собой.