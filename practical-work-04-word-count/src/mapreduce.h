#ifndef MAPREDUCE_H
#define MAPREDUCE_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>

// Key-Value pair
typedef struct {
    char *key;
    char *value;
} KeyValue;

// Key-Value list
typedef struct KeyValueList {
    KeyValue kv;
    struct KeyValueList *next;
} KeyValueList;

// Function pointer types
typedef void (*MapFunc)(const char *key, const char *value, KeyValueList **output);
typedef void (*ReduceFunc)(const char *key, const char **values, int num_values, FILE *output);

// MapReduce context
typedef struct {
    int num_mappers;
    int num_reducers;
    char *input_dir;
    char *output_dir;
    MapFunc map_func;
    ReduceFunc reduce_func;
    KeyValueList **intermediate;  // [reducer_id][]
    pthread_mutex_t *mutexes;     // Mutexes for each reducer
} MapReduceContext;

// MapReduce functions
int mapreduce_init(MapReduceContext *ctx, int num_mappers, int num_reducers,
                   const char *input_dir, const char *output_dir,
                   MapFunc map_func, ReduceFunc reduce_func);

int mapreduce_run(MapReduceContext *ctx);

void mapreduce_cleanup(MapReduceContext *ctx);

// Hash function for partitioning
int hash_partition(const char *key, int num_reducers);

// Helper functions
void emit_intermediate(MapReduceContext *ctx, const char *key, const char *value);
void emit_final(FILE *output, const char *key, const char *value);

#endif // MAPREDUCE_H

