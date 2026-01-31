package internal

import "strings"

type Router struct {
	root *RouterNode
}

type RouterNode struct {
	pathPart     string
	children     map[string]*RouterNode
	wildCardNode *RouterNode
	varNode      *RouterNode
	data         RequestPipeline
}

func NewRouter() *Router {
	return &Router{
		root: newNode(""),
	}
}

func (self *Router) setupPath(path string, handler RequestPipeline) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := self.root

	for _, part := range parts {

		if part == "*" {
			if current.wildCardNode == nil {
				current.wildCardNode = newNode(part)
			}
			current = current.wildCardNode
			continue
		}

		if strings.HasPrefix(part, ":") {
			if current.varNode == nil {
				current.varNode = newNode(part)
			}
			current = current.varNode
			continue
		}

		if _, exists := current.children[part]; !exists {
			current.children[part] = newNode(part)
		}
		current = current.children[part]
	}
	current.data = handler
}

func (self *Router) EvalRoute(path string) (RequestPipeline, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	type stackFrame struct {
		node      *RouterNode
		partIndex int
		state     int // 0: to visit static, 1: to visit var, 2: to visit wildcard
	}

	stack := []stackFrame{{node: self.root, partIndex: 0, state: 0}}

	for len(stack) > 0 {
		frame := &stack[len(stack)-1]

		if frame.partIndex == len(parts) {
			if frame.node.data != (RequestPipeline{}) {
				return frame.node.data, true
			}
			stack = stack[:len(stack)-1]
			continue
		}

		part := parts[frame.partIndex]

		switch frame.state {
		case 0:
			frame.state = 1
			if child, exists := frame.node.children[part]; exists {
				stack = append(stack, stackFrame{node: child, partIndex: frame.partIndex + 1, state: 0})
			}
		case 1:
			frame.state = 2
			if frame.node.varNode != nil {
				stack = append(stack, stackFrame{node: frame.node.varNode, partIndex: frame.partIndex + 1, state: 0})
			}
		case 2:
			frame.state = 3
			if frame.node.wildCardNode != nil {
				stack = append(stack, stackFrame{node: frame.node.wildCardNode, partIndex: frame.partIndex + 1, state: 0})
			}
		case 3:
			stack = stack[:len(stack)-1]
		}
	}

	return RequestPipeline{}, false
}

func newNode(part string) *RouterNode {
	return &RouterNode{
		pathPart: part,
		children: make(map[string]*RouterNode),
		data: RequestPipeline{},
	}
}