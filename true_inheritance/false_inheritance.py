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


team = [
    Employee("Alice", 50_000),
    Manager("Bob", 60_000, 10_000),
    Director("Carol", 70_000, 15_000, 20_000),
]

for w in team:
    print(f"{w.role():<12} salary: {w.salary():.0f}")
# Employee     salary: 50000
# Manager      salary: 70000
# Director     salary: 105000