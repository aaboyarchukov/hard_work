# Истинное наследование

## Все что вы знали об ООП - неверно
---
Суть такова - необходимо понимать суть решения проблем с помощью паттернов проектирования, тогда все паттерны будут сводиться к одному - Visitor, который можно использовать практически во всех случаях.

## Stop Memorizing Design Patterns: Use This Desicion Tree Instead
---
В данной статье описывается способ изучения паттернов не из заучивания списка возможных паттернов, а через поиск по дереву, отталкиваясь от проблемы.
То есть изначально необходимо определить проблему, с которой вы сталкиваетесь, а затем спускаясь по дереву искать подходящие пути решения.

Дерево выбора:

[design_patterns_decision_tree.html](https://github.com/aaboyarchukov/hard_work/blob/main/true_inheritance/design_patterns_decision_tree.html)

## Пример
---
Пример показан на python, так как на Go нет классического наследования

[false_inheritance.py](https://github.com/aaboyarchukov/hard_work/blob/main/true_inheritance/false_inheritance.py)

```python
class Employee:
    def __init__(self, name: str, base_salary: float):
        self.name = name
        self.base_salary = base_salary

    def salary(self) -> float:
        return self.base_salary

    def role(self) -> str:
        return "Employee"


class Manager(Employee):
    def __init__(self, name: str, base_salary: float, bonus: float):
        super().__init__(name, base_salary)
        self.bonus = bonus

    def salary(self) -> float:
        return super().salary() + self.bonus

    def role(self) -> str:
        return "Manager"


class Director(Manager):
    def __init__(self, name: str, base_salary: float, bonus: float, stock_options: float):
        super().__init__(name, base_salary, bonus)
        self.stock_options = stock_options

    def salary(self) -> float:
        return super().salary() + self.stock_options

    def role(self) -> str:
        return "Director"
```

[true_inheritance.py](https://github.com/aaboyarchukov/hard_work/blob/main/true_inheritance/false_inheritance.py)

```python
from abc import ABC, abstractmethod

class SalaryVisitor(ABC):
    @abstractmethod
    def regular_employee_salary(emp : 'Employee') -> float: ...
    
    @abstractmethod
    def manager_salary(m : 'Manager') -> float: ...

    @abstractmethod
    def director_salary(d : 'Director') -> float: ...

class EmployeeVisitor(SalaryVisitor):
    def regular_employee_salary(emp : 'Employee') -> float:
        return emp.base_salary
    
    def manager_salary(m : 'Manager') -> float:
        return m.base_salary + m.bonus

    def director_salary(d : 'Director') -> float:
        return d.base_salary + d.bonus + d.stock_options

class Employee:
    def __init__(self, name: str, base_salary: float):
        self.name = name
        self.base_salary = base_salary
    
    def accept_salary(self, v : 'SalaryVisitor'):
        v.regular_employee_salary(self)


class Manager(Employee):
    def __init__(self, name: str, base_salary: float, bonus: float):
        super().__init__(name, base_salary)
        self.bonus = bonus
    
    def accept_salary(self, v : 'SalaryVisitor'):
        v.manager_salary(self)


class Director(Manager):
    def __init__(self, name: str, base_salary: float, bonus: float, stock_options: float):
        super().__init__(name, base_salary, bonus)
        self.stock_options = stock_options
    
    def accept_salary(self, v : 'SalaryVisitor'):
        v.director_salary(self)
```

***Что я сделал***: Мы завели интефейс "Посетителя", который добавляет метод по получению зарплаты в зависимости от должности. Таким образом мы ушли от переопределения метода `salary` в каждом классе иерархии, соответственно перешли к истинному наследованию. Но все же нам также необходимо было определять новый метод, который определяет новую операцию над классами.

***Чего мы добились***: уменьшили связность между классами, при добавлении нового класса в иерархии достаточно определить новые операции только для него; при изменении операции в родительских классах, ничего не надо менять в дочерних; также код классов уменьшился; также при изменении текущих операций, понадобится изменить лишь самого посетителя

**В каких случаях использовать**: в случаях, когда у нас постоянная иерархия и надо добавлять новые операции над ними

***Резюме***: считаю, что результат получился чуть лучше, так как уменшили связность и количество кода в самих классах, но код стал чуть менее читаем, а также считаю, что мой пример не такой показательный, лучше получится в тех случаях, когда у нас объекты существуют в определенной иерархии/структуре, и им надо определить новые операции

