#include "mapreduce.h"
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>
#include <ctype.h>

// Hash function for partitioning keys to reducers
int hash_partition(const char *key, int num_reducers) {
    unsigned int hash = 0;
    for (int i = 0; key[i] != '\0'; i++) {
        hash = hash * 31 + (unsigned char)key[i];
    }
    return hash % num_reducers;
}

// Emit intermediate key-value pair
void emit_intermediate(MapReduceContext *ctx, const char *key, const char *value) {
    int partition = hash_partition(key, ctx->num_reducers);
    
    pthread_mutex_lock(&ctx->mutexes[partition]);
    
    KeyValueList *new_node = malloc(sizeof(KeyValueList));
    new_node->kv.key = strdup(key);
    new_node->kv.value = strdup(value);
    new_node->next = ctx->intermediate[partition];
    ctx->intermediate[partition] = new_node;
    
    pthread_mutex_unlock(&ctx->mutexes[partition]);
}

// Emit final key-value pair
void emit_final(FILE *output, const char *key, const char *value) {
    fprintf(output, "%s\t%s\n", key, value);
}

// Mapper thread function
typedef struct {
    MapReduceContext *ctx;
    char *input_file;
} MapperArgs;

void *mapper_thread(void *arg) {
    MapperArgs *args = (MapperArgs *)arg;
    MapReduceContext *ctx = args->ctx;
    FILE *file = fopen(args->input_file, "r");
    
    if (file == NULL) {
        fprintf(stderr, "Error: Cannot open file %s\n", args->input_file);
        return NULL;
    }
    
    char line[4096];
    while (fgets(line, sizeof(line), file)) {
        // Remove newline
        line[strcspn(line, "\n")] = '\0';
        
        // Call map function
        // For word count, key is filename, value is line
        ctx->map_func(args->input_file, line, &ctx->intermediate[0]);
    }
    
    fclose(file);
    free(args->input_file);
    free(args);
    return NULL;
}

// Reducer thread function
typedef struct {
    MapReduceContext *ctx;
    int reducer_id;
} ReducerArgs;

void *reducer_thread(void *arg) {
    ReducerArgs *args = (ReducerArgs *)arg;
    MapReduceContext *ctx = args->ctx;
    
    // Create output file
    char output_file[256];
    snprintf(output_file, sizeof(output_file), "%s/part-%05d", 
             ctx->output_dir, args->reducer_id);
    
    FILE *output = fopen(output_file, "w");
    if (output == NULL) {
        fprintf(stderr, "Error: Cannot create output file %s\n", output_file);
        return NULL;
    }
    
    // Collect all key-value pairs for this reducer
    KeyValueList *list = ctx->intermediate[args->reducer_id];
    
    // Group by key
    // Simple implementation: collect all, then group
    KeyValueList *current = list;
    while (current != NULL) {
        // Find all values for this key
        const char *key = current->kv.key;
        char **values = NULL;
        int num_values = 0;
        int capacity = 10;
        
        values = malloc(capacity * sizeof(char*));
        
        KeyValueList *iter = list;
        while (iter != NULL) {
            if (strcmp(iter->kv.key, key) == 0) {
                if (num_values >= capacity) {
                    capacity *= 2;
                    values = realloc(values, capacity * sizeof(char*));
                }
                values[num_values++] = iter->kv.value;
            }
            iter = iter->next;
        }
        
        // Call reduce function
        ctx->reduce_func(key, (const char **)values, num_values, output);
        
        // Move to next unique key
        while (current != NULL && strcmp(current->kv.key, key) == 0) {
            current = current->next;
        }
        
        free(values);
    }
    
    fclose(output);
    free(args);
    return NULL;
}

// Initialize MapReduce context
int mapreduce_init(MapReduceContext *ctx, int num_mappers, int num_reducers,
                   const char *input_dir, const char *output_dir,
                   MapFunc map_func, ReduceFunc reduce_func) {
    ctx->num_mappers = num_mappers;
    ctx->num_reducers = num_reducers;
    ctx->input_dir = strdup(input_dir);
    ctx->output_dir = strdup(output_dir);
    ctx->map_func = map_func;
    ctx->reduce_func = reduce_func;
    
    // Allocate intermediate storage
    ctx->intermediate = calloc(num_reducers, sizeof(KeyValueList*));
    ctx->mutexes = malloc(num_reducers * sizeof(pthread_mutex_t));
    
    for (int i = 0; i < num_reducers; i++) {
        pthread_mutex_init(&ctx->mutexes[i], NULL);
    }
    
    // Create output directory
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "mkdir -p %s", output_dir);
    system(cmd);
    
    return 0;
}

// Run MapReduce job
int mapreduce_run(MapReduceContext *ctx) {
    // Phase 1: Map
    DIR *dir = opendir(ctx->input_dir);
    if (dir == NULL) {
        fprintf(stderr, "Error: Cannot open input directory %s\n", ctx->input_dir);
        return -1;
    }
    
    struct dirent *entry;
    pthread_t *mapper_threads = malloc(ctx->num_mappers * sizeof(pthread_t));
    int file_count = 0;
    char **input_files = NULL;
    int capacity = 10;
    input_files = malloc(capacity * sizeof(char*));
    
    // Collect input files
    while ((entry = readdir(dir)) != NULL) {
        if (entry->d_type == DT_REG) {  // Regular file
            if (file_count >= capacity) {
                capacity *= 2;
                input_files = realloc(input_files, capacity * sizeof(char*));
            }
            
            char full_path[512];
            snprintf(full_path, sizeof(full_path), "%s/%s", 
                     ctx->input_dir, entry->d_name);
            input_files[file_count++] = strdup(full_path);
        }
    }
    closedir(dir);
    
    printf("Found %d input files\n", file_count);
    
    // Launch mapper threads
    for (int i = 0; i < file_count && i < ctx->num_mappers; i++) {
        MapperArgs *args = malloc(sizeof(MapperArgs));
        args->ctx = ctx;
        args->input_file = input_files[i];
        
        pthread_create(&mapper_threads[i], NULL, mapper_thread, args);
    }
    
    // Wait for mappers
    for (int i = 0; i < file_count && i < ctx->num_mappers; i++) {
        pthread_join(mapper_threads[i], NULL);
    }
    
    free(mapper_threads);
    for (int i = 0; i < file_count; i++) {
        free(input_files[i]);
    }
    free(input_files);
    
    printf("Map phase completed\n");
    
    // Phase 2: Reduce
    pthread_t *reducer_threads = malloc(ctx->num_reducers * sizeof(pthread_t));
    
    for (int i = 0; i < ctx->num_reducers; i++) {
        ReducerArgs *args = malloc(sizeof(ReducerArgs));
        args->ctx = ctx;
        args->reducer_id = i;
        
        pthread_create(&reducer_threads[i], NULL, reducer_thread, args);
    }
    
    // Wait for reducers
    for (int i = 0; i < ctx->num_reducers; i++) {
        pthread_join(reducer_threads[i], NULL);
    }
    
    free(reducer_threads);
    
    printf("Reduce phase completed\n");
    
    return 0;
}

// Cleanup MapReduce context
void mapreduce_cleanup(MapReduceContext *ctx) {
    // Free intermediate data
    for (int i = 0; i < ctx->num_reducers; i++) {
        KeyValueList *current = ctx->intermediate[i];
        while (current != NULL) {
            KeyValueList *next = current->next;
            free(current->kv.key);
            free(current->kv.value);
            free(current);
            current = next;
        }
        pthread_mutex_destroy(&ctx->mutexes[i]);
    }
    
    free(ctx->intermediate);
    free(ctx->mutexes);
    free(ctx->input_dir);
    free(ctx->output_dir);
}

