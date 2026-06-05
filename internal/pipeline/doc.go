// Package pipeline turns validated KDL specs into laid-out diagrams.
//
// Architecture (source of truth chain):
//
//	KDL file
//	  → spec.LoadFile + spec.Validate   (syntax + referential integrity)
//	  → pipeline.BuildDeck              (geometry: positions, routes, sizes)
//	  → scene.Evaluate per slide        (semantics: stack, rules, score, paint order)
//	  → validate / advise / describe / render (consumers of pipeline.Result)
//
// Geometry lives in model.Diagram. Visual semantics live in scene.Evaluation.
// Render projects Diagram → pixels using paint order defined by the engine.
package pipeline
