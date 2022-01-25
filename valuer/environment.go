package valuer

type Environment struct {
	Values    map[string]Value
	Enclosing *Environment
}

func (env *Environment) Define(key string, v Value) {
	env.Values[key] = v
}

func (env *Environment) Get(key string) (Value, bool) {
	if v, ok := env.Values[key]; ok {
		return v, true
	}
	if env.Enclosing != nil {
		return env.Enclosing.Get(key)
	}
	return nil, false
}

func (env *Environment) GetAt(distance int, key string) (Value, bool) {
	return env.ancestor(distance).Get(key)
}

func (env *Environment) Assign(key string, v Value) bool {
	if _, ok := env.Values[key]; ok {
		env.Values[key] = v
		return true
	}
	if env.Enclosing != nil {
		return env.Enclosing.Assign(key, v)
	}
	return false
}

func (env *Environment) AssignAt(distance int, key string, v Value) bool {
	return env.ancestor(distance).Assign(key, v)
}

func (env *Environment) ancestor(distance int) *Environment {
	cur := env
	for i := 0; i < distance; i++ {
		cur = cur.Enclosing
	}
	return cur
}

func NewEnvironment() *Environment {
	return &Environment{
		Values: make(map[string]Value),
	}
}

func NewEnclosing(env *Environment) *Environment {
	closing := NewEnvironment()
	closing.Enclosing = env
	return closing
}
