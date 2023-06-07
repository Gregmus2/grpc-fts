package config

import (
	"errors"
	"fmt"
)

type Tag int

const (
	ToDo Tag = iota
	InProgress
	Done
)

type Node struct {
	tag    Tag
	entity TestCase
}

func Sort(entities []TestCase) ([]TestCase, error) {
	nodes := make(map[string]*Node, len(entities))
	for _, entity := range entities {
		nodes[entity.Name] = &Node{tag: ToDo, entity: entity}
	}

	result := make([]TestCase, 0, len(entities))
	for _, node := range nodes {
		testCases, err := sortNode(node, nodes)
		if err != nil {
			return nil, err
		}

		result = append(result, testCases...)
	}

	return result, nil
}

func sortNode(node *Node, nodes map[string]*Node) ([]TestCase, error) {
	switch node.tag {
	case Done:
		return []TestCase{}, nil
	case InProgress:
		return nil, errors.New("cycle was found")
	case ToDo:
		result := make([]TestCase, 0)
		node.tag = InProgress
		for _, name := range node.entity.DependsOn {
			n, ok := nodes[name]
			if !ok {
				return nil, errors.New("node not found")
			}

			entities, err := sortNode(n, nodes)
			if err != nil {
				return nil, err
			}

			result = append(result, entities...)
		}

		node.tag = Done
		result = append(result, node.entity)

		return result, nil
	default:
		return nil, fmt.Errorf("unknown tag %d", node.tag)
	}
}
