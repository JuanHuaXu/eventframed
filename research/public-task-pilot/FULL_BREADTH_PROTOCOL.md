# Derived presence and full-breadth diagnosis

Two fresh6400-record768d graphs, same bulk construction. Check each query's
vector directly by collection.Get, separate from the owned map and ANN. Compare
default400 with requested6400 on the same graph for all1024 self queries.
Record candidates, errors, stored-vector equality, default-adapter agreement and
diagnostic timing. Read stored vectors before both arms; timings are not directly
comparable to prior experiments because this warms storage. No serving claim.

At ef equal to record count a graph traversal still need not visit disconnected
nodes. Do not call it exhaustive exact search. A remaining miss with exact stored
vector rules out missing stored payload in THIS run, but not absent index nodes,
disconnected topology, distance-provider errors or traversal defects.

Discovery only. Preserve prior artifacts. No backend patch or production change.
