import parser
import cmath

max_val = 64
def escapeParse(c, formula):
    z = c
    for i in range(max_val):
        if abs(z) > 2:
            return True
        z = eval(parser.expr(formula).compile())
    return False

def escape(c):
    z = c
    for i in range(max_val):
        if abs(z) > 2:
            return True
        z = z ** 2 + c
    return False

cmplx = complex(1, .01)
print(escape(cmplx))
print(escapeParse(cmplx, "z ** 2 + c"))