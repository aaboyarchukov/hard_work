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


team = [
    Employee("Alice", 50_000),
    Manager("Bob", 60_000, 10_000),
    Director("Carol", 70_000, 15_000, 20_000),
]

for w in team:
    print(f"salary: {w.accept_salary(SalaryVisitor()):.0f}")
# Employee     salary: 50000
# Manager      salary: 70000
# Director     salary: 105000