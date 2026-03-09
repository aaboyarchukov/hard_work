# Минимизация шагов

В данном уроке нам необходимо научиться вносить изменения в разработку своего проекта/решения задачи, за счет небольших шагов, которые фильтруются через прогон тестов. Основная идея заключается в следующем:
- первым делом мы разрабатываем спецификацию к данной задаче
- далее по спецификации мы начинаем разрабатывать тесты + ставим заглушки (или минимальные версии, которые имплементируют наш функционал)
	- тесты которые мы разрабатываем покрывают минимум кейсов, ведь для начала мы достигаем той точки - когда у нас все тесты проходят
- далее, после реализации базового минимума, мы наращиваем обороты -> добавляем кейсы в наш тест, а затем рефакторим функционал до тех пор, пока тесты снова не будут зелеными

То есть алгоритм следующий:

1. Напиши **один** тест (можно даже часть теста), сделай так чтобы он прошёл — хоть фейком.
2. Закоммить.
3. Замени фейк на кусочек настоящей реализации.
4. Закоммить.
5. Добавь следующий тест — и снова по кругу.

То есть у нас идет циклическая итерация по разработке. Таким образом мы не держим большой контекст в голове, а максимально его минимизиуем!

Для простоты реализации и написания тестов - возьмем алгоритмические задачи.
## Примеры

---
### Пример 1

---
### Спецификация

---

```
Задача:

Необходимо найти самую длинную монотонно-возрастающую подпоследовательность в последовательности

[7, 1, 2, 3, 0, 4, 5, 6, 5] -> [1, 2, 3, 4, 5, 6]

Алгоритм:

- необходимо определить массив (collection), в котором будут хранится упорядоченные элементы, длина этого массива - максимальная длина искомой последовательности
- далее при итерации исходноо массива, мы ищем позицию для вставки в collection, ищем за счет бинарного поиска, но который ищет такое место в массиве, чтобы он сохранил свою упорядоченность
- таким образом у нас формируется массив упорядоченных элементов, которые формируют последовательность, осталось их соотнести между собой по местам с изначальным массивов, делаем мы это с помощью еще ожного массива, который хранит в себе индексы родителя каждого элемента collection, для того, чтобы восставновить его в конце
- после сбора целевого массива с последовательностью, а такэе имея индексы родителей - мы можем просто восстановить нужную нам последовательность
  

Но как по мне данное решение еще можно оптимизировать 
```

---
### [example1.go](https://github.com/aaboyarchukov/hard_work/blob/main/minimize_steps/example1/solution.go)

---

```go
func StrictlyMonotonousSequence(array []int) []int {
	size := len(array)
	if size <= 1 {
		return array
	}

	collection := make([]int, 0, size)
	parents := make([]int, size)

	for indx := range array {
		target := array[indx]
		collection_pos := example2.BinarySearchLeft(collection, target)

		if collection_pos == len(collection) {
			collection = append(collection, target)
		} else {
			collection[collection_pos] = target
		}

		if collection_pos > 0 {
			parents[collection_pos] = collection[collection_pos-1]
		} else {
			parents[collection_pos] = -1
		}
	}

	collection_size := len(collection)
	result := make([]int, 0, collection_size)

	result = append(result, collection[collection_size-1])
	for indx := collection_size - 1; indx > 0; indx-- {
		result = append(result, parents[indx])
	}

	slices.Reverse(result)

	return result
}
```

---
### [example1_test.go](https://github.com/aaboyarchukov/hard_work/blob/main/minimize_steps/example1/solution_test.go)

---

```go
package example1

import (
	"slices"
	"testing"
)

func TestStrictlyMonotonousSequence(t *testing.T) {
	type strictlyMonotonousSequenceCase struct {
		Name   string
		Input  []int
		Output []int
	}

	cases := []strictlyMonotonousSequenceCase{
		// ...
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(subT *testing.T) {
			result := StrictlyMonotonousSequence(testCase.Input)
			if slices.Compare(result, testCase.Output) != 0 {
				subT.Fatalf("FAILED: %s, wanted: %v, got: %v", testCase.Name, testCase.Output, result)
			}
		})
	}
}

```

---
### История коммитов

---

```bash
515a6c0 (HEAD -> main, origin/main, origin/HEAD) minimize_steps | example 1 - refactor solution - tests passed
fd196d3 minimize_steps | example 1 - refactor solution - tests passed
0c71ca4 minimize_steps | example 1 - refactor solution - tests passed
06b9720 minimize_steps | example 1 - refactor solution - tests passed
dcbc82e minimize_steps | example 1 - refactor solution - tests passed
b509010 minimize_steps | example 1 - first attemp, init solution and first test case
```

---

### Рефлексия по решению

---

Задача непростая и сложно по началу наработанное решение откатывать назад, ведь оно для тебя что-то значило, так как ты потратил на него время и силы. И понял, что такое чувство могло возникнуть по тому, что сами изменения очень большие. 
Тогда перешел на еще более маленькие изменения, чтобы сохранять правильные шаги решения, а также с большей легкостью расставаться с изменениями.

---
### Пример 2

---
### Спецификация

---

```
Задача:

Создать функцию биарного поиска, которая ищет место в массиве таким образом, чтобы после вставки - он оставаля отсортированным.

Алгоритм:

- циклом указателями проходимся с концов массива, пока они не станут равны, как только они стали равны - найдено нужное место
  
нужным местом является либо слева от элемента равного найденному, либо при вставке он должен сохранять упорядоченность (делается это посредством сдвига указателя от середины диапозона как в классическом бинарном поиске) 
```

---
### [example2.go](https://github.com/aaboyarchukov/hard_work/blob/main/minimize_steps/example2/solution.go)

---

```go
func BinarySearchLeft(array []int, target int) int {
	left, right := 0, len(array)

	for left < right {
		middle := (left + right) / 2

		if target > array[middle] {
			left = middle + 1
		}

		if target <= array[middle] {
			right = middle
		}

	}

	return left
}
```

---
### [example2_test.go](https://github.com/aaboyarchukov/hard_work/blob/main/minimize_steps/example2/solution_test.go)

---

```go
package example2

import "testing"

func TestBinarySearchLeft(t *testing.T) {
	type binarySearchLeftCase struct {
		Name       string
		InputArray []int
		Target     int
		Output     int
	}

	cases := []binarySearchLeftCase{
		// ...
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(subT *testing.T) {
			resultIndx := BinarySearchLeft(testCase.InputArray, testCase.Target)

			if resultIndx != testCase.Output {
				subT.Fatalf("FAILED: %s, wanted: %v, got: %d", testCase.Name, testCase.Output, resultIndx)
			}
		})
	}
}

```

---
### История коммитов

---

```bash
8fb2a93 (HEAD -> main, origin/main, origin/HEAD) minimize_steps | example 2 - refactor solution and tests - tests passed
b4eaf99 minimize_steps | example 2 - refactor solution - tests passed
23c3da0 minimize_steps | example 2 - refactor solution and tests - tests passed
a336c26 minimize_steps | example 2 - init solutionand tests - tests passed
```

---

### Рефлексия по решению

---

С данной задачей было полегче, необходимо было реализовать поиск для предыдущей задачи, делил на максимально мелкие шаги, чтобы не стирать большие исправления кодовой базы. Сначала писал один тест-кейс, проверял его, затем коммитил при успехе, далее рефакторил и создавал новый тест-кейс, и так в цикле, пока не покрою все случаи. 

---

## Итог

Проанализировав данный способ на использовании при решении задач понял, что он может быть весьма эффективен, так как заставляет задуматься о мимнимизации исправлений, а это способствует тому, что ты держишь в голове меньший контекст, значит легче становится проводить рефакторинг, так как ты работаешь точечно.

Простой принцип "Разделяй и властвуй", который используется через TDD.

Единственным недостатком заметил, что при таком принципе, ты можешь иногда при решении испытывать неприятное ощещение, возникающее при откате изменений, если в изменениях были простые инициализации, которые пришли в процессе рефакторинга. Но думаю, что здесь стоит оптимизировать модули, которые ты рефакторишь, уменьшая (в контексте объема рефакторинга) их до такой степени, что инициализация данных или другая обвязка не повлияют при прогоне тестов и не пропадут при откате, если тесты были провалены.