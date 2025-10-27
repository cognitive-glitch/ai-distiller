"""
Edge case: Deeply nested classes and functions.
Tests parser's recursion handling.
"""


class Level1:
    """First level class."""

    class Level2:
        """Second level nested class."""

        class Level3:
            """Third level nested class."""

            class Level4:
                """Fourth level nested class."""

                class Level5:
                    """Fifth level nested class."""

                    def deeply_nested_method(self):
                        """Method in deeply nested class."""

                        def inner_function_1():
                            """First level inner function."""

                            def inner_function_2():
                                """Second level inner function."""

                                def inner_function_3():
                                    """Third level inner function."""

                                    def inner_function_4():
                                        """Fourth level inner function."""

                                        def inner_function_5():
                                            """Fifth level inner function."""
                                            return "deeply nested"

                                        return inner_function_5()

                                    return inner_function_4()

                                return inner_function_3()

                            return inner_function_2()

                        return inner_function_1()


# Deeply nested control structures
def complex_nesting():
    """Function with deeply nested control structures."""
    if True:
        if True:
            if True:
                if True:
                    if True:
                        if True:
                            if True:
                                if True:
                                    if True:
                                        if True:
                                            return "10 levels deep"
