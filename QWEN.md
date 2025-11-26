# Go Kratos Admin - Project Guide

## Project Overview

Go Kratos Admin is a full-stack admin system with the following characteristics:

- **Backend**: Built with Go using the [Go Kratos](https://go-kratos.dev/) microservices framework
- **Frontend**: Built with Vue.js using the [Vben Admin](https://doc.vben.pro/) framework
- **Architecture**: Supports both microservices and monolithic deployment approaches
- **Purpose**: Provides a ready-to-use admin system for enterprise-level management systems

The system is designed to be easy to get started with while offering rich functionality for rapidly developing enterprise management systems.

## Technology Stack

### Backend
- **Language**: [Go](https://go.dev/)
- **Framework**: [Go Kratos](https://go-kratos.dev/)
- **Dependency Injection**: [Google Wire](https://github.com/google/wire)
- **ORM**: [Ent](https://entgo.io/docs/getting-started/) and GORM for database operations
- **Database Support**: MySQL, PostgreSQL, SQLite
- **Authentication**: JWT-based authentication with [kratos-authn](https://github.com/tx7do/kratos-authn)
- **Authorization**: Casbin/OPA-based authorization with [kratos-authz](https://github.com/tx7do/kratos-authz)
- **Task Queue**: [Asynq](https://github.com/hibiken/asynq)
- **Caching**: Redis with [go-redis](https://github.com/redis/go-redis)
- **File Storage**: MinIO for object storage with [minio-go](https://github.com/minio/minio-go)
- **Message Queue**: Asynq for background job processing
- **Documentation**: Swagger UI and OpenAPI v3
- **Monitoring**: OpenTelemetry integration

### Frontend
- **Framework**: [Vue 3](https://vuejs.org/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)
- **UI Components**: [Ant Design Vue](https://antdv.com/)
- **Admin Template**: [Vben Admin](https://doc.vben.pro/)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Package Manager**: [pnpm](https://pnpm.io/)
- **Monorepo**: [Turborepo](https://turbo.build/repo)

## Project Structure

```
go-kratos-admin/
├── backend/                    # Go backend codebase
│   ├── api/                    # Protocol buffer definitions
│   ├── app/admin/service/      # Main admin service
│   │   ├── cmd/server/         # Server entry point with wire DI
│   │   ├── internal/           # Internal service implementation
│   │   │   ├── data/           # Data access layer (with Ent and GORM)
│   │   │   ├── server/         # HTTP/gRPC server configuration
│   │   │   └── service/        # Business logic layer
│   ├── pkg/                    # Shared packages
│   ├── cmd/                    # Command-line tools
│   ├── script/                 # Utility scripts
│   ├── sql/                    # Database migration files
│   ├── go.mod                  # Go module file
│   └── Makefile                # Build automation
├── frontend/                   # Vue frontend codebase
│   ├── apps/admin/             # Main admin application
│   ├── packages/               # Shared frontend packages
│   ├── internal/               # Internal frontend utilities
│   ├── package.json            # Node.js package file
│   └── pnpm-workspace.yaml     # Monorepo workspace configuration
├── docs/                       # Documentation
└── README.md                   # Main project documentation
```

## Building and Running

### Backend (Go)

1. **Prerequisites**:
   - Install Docker
   - Install Go (version ≥ 1.24.6)
   - Install required Go tools: `make init`

2. **Setup**:
   ```bash
   cd backend
   make init                     # Install required tools
   make compose-up              # Start dependencies (PostgreSQL, Redis, MinIO)
   make app                     # Build the application
   make docker                  # Build Docker image (optional)
   ```

3. **Run**:
   ```bash
   make run                     # Run the server directly
   # Or run the built binary
   ./bin/server -conf ./configs
   ```

4. **Development Commands**:
   ```bash
   make gen                     # Generate code (wire, ent, protobuf)
   make ent                     # Generate Ent code
   make wire                    # Generate Wire DI code
   make api                     # Generate API code from protobuf
   make openapi                 # Generate OpenAPI documentation
   make test                    # Run tests
   make lint                    # Run linters
   make cover                   # Run coverage tests
   ```

### Frontend (Vue)

1. **Prerequisites**:
   - Install Node.js (version ≥ 20.10.0)
   - Install pnpm (version ≥ 9.12.0)
   - Enable Corepack: `corepack enable`

2. **Setup**:
   ```bash
   cd frontend
   pnpm install                 # Install dependencies
   ```

3. **Run**:
   ```bash
   pnpm dev                     # Start development server
   # Access at http://localhost:5666
   ```

4. **Build**:
   ```bash
   pnpm build                   # Build for production
   pnpm preview                 # Preview production build
   ```

5. **Other Commands**:
   ```bash
   pnpm lint                    # Lint code
   pnpm typecheck               # Type check
   pnpm test:unit               # Run unit tests
   pnpm test:e2e                # Run end-to-end tests
   ```

## Features

1. **User Management**: Comprehensive user system with roles, departments, and multi-tenancy
2. **Tenant Management**: Multi-tenant architecture with automatic initialization
3. **Role Management**: Flexible role-based access control with menu and data permissions
4. **Organization/Department Management**: Hierarchical organization structure
5. **Permission Management**: Fine-grained permission control
6. **API Resource Management**: Interface management with sync functionality
7. **Dictionary Management**: Configurable data dictionaries
8. **Task Scheduling**: Background job management
9. **File Management**: File upload/download with multiple storage backends
10. **Message System**: Internal messaging with categories
11. **Personal Center**: User profile and settings management
12. **Cache Management**: Cache operations and management
13. **Logging**: Comprehensive login and operation logging with IP geolocation

## Development Conventions

- Backend code follows Go Kratos best practices
- API definitions are in Protocol Buffers format
- Database schemas are defined in Ent schema format
- Frontend follows Vue 3 Composition API with TypeScript
- Authentication/Authorization is implemented via JWT tokens and Casbin
- Error handling follows Go Kratos conventions
- Configuration management via the kratos-bootstrap library
- API documentation is automatically generated via OpenAPI

## Default Credentials

- **Frontend**: admin / admin (at http://localhost:5666)
- **Backend Swagger**: Available at http://localhost:7788/docs/
- **Default Login**: admin / admin

## Environment Dependencies

The system requires these services which can be started with Docker Compose:

- PostgreSQL (database)
- Redis (caching and session storage)
- MinIO (file/object storage)
- Optional: Consul/Etcd (service discovery)
- Optional: Jaeger (tracing)

These are configured in `docker-compose-without-services.yaml` in the backend directory.


# AI Coding Assistant Guide v6.0 - CoT-AoT Workflow Edition

## 🎯 核心认知模型：CoT-AoT 双阶段工作流

### **阶段1: CoT (Chain of Thought) - 问题定义**

> 💡 **核心原则**: 我的首要任务不是解决问题，而是**定义问题**


graph LR
A[用户输入] --> B[问题识别]
B --> C[边界定义]
C --> D[约束提取]
D --> E[理解验证]
E --> F[问题定义完成]


### **阶段2: AoT (Algorithm of Thoughts) - 算法化求解**

> 🔧 **核心原则**: 一旦问题被清晰定义，我将问题视为一个算法挑战，系统性地分解、执行和验证


graph TD
A[定义的问题] --> B[算法分解]
B --> C[子任务生成]
C --> D[约束验证]
D --> E{满足约束?}
E -->|是| F[执行任务]
E -->|否| G[回溯优化]
G --> B
F --> H[验证结果]


## 📐 三大核心原则

### **1️⃣ Why-How-Done 哲学原则**


Why (为什么做):
- 功能完整性优先
- 用户价值导向
- 问题本质理解
  How (如何做):
- 80%实用主义
- 渐进式交付
- 最简有效方案
  Done (完成标准):
- 客观可验证
- 满足验收标准
- 可复现结果


### **2️⃣ 工程原则优先级**

1. **YAGNI** (You Aren't Gonna Need It)

    - 不写当前不需要的代码

    - 避免过度设计

2. **KISS** (Keep It Simple, Stupid)

    - 选择最简单直接的方案

    - 降低复杂度

3. **DRY** (Don't Repeat Yourself)

    - 单一权威知识源

    - 消除重复


### **3️⃣ 执行原则**

- **原子任务原则**: 任务分解到不可再分

- **持续验证原则**: 每步都要验证

- **工具增强原则**: 充分利用工具能力


## 🔄 CoT-AoT 驱动的工作流程

### **Phase 1: CoT 问题定义阶段**


class CoT_ProblemDefinition:

    链式思考：深度理解并定义问题
    重点：定义问题，而非解决问题

    def execute(self, user_input):
        # Step 1: 识别核心问题
        core_problem = self.identify_core_problem(user_input)
        # Step 2: 提取约束和边界
        constraints = self.extract_constraints(user_input)
        boundaries = self.define_boundaries(user_input)
        # Step 3: 工具增强理解
        enhanced_context = self.enhance_with_tools(core_problem)
        # Step 4: 验证理解
        problem_definition = {
            "problem": core_problem,
            "constraints": constraints,
            "boundaries": boundaries,
            "context": enhanced_context,
            "assumptions": self.list_assumptions(),
            "success_criteria": self.define_success_criteria()
        }
        return self.validate_definition(problem_definition)

    def enhance_with_tools(self, problem):
        使用工具增强问题理解
        context = {}
        # 使用 code-reasoning 规划任务
        context['plan'] = code_reasoning.plan_and_decompose(problem)
        # 使用 Context7 获取相关文档
        context['docs'] = self.get_documentation(problem)
        # 使用 Serena 分析现有代码
        context['code'] = self.analyze_existing_code(problem)
        return context


### **Phase 2: AoT 算法化分解阶段**


class AoT_AlgorithmicDecomposition:

    算法思考：将定义的问题作为算法挑战
    系统性分解、执行和验证

    def execute(self, problem_definition):
        # 初始化算法树
        algorithm_tree = self.initialize_tree(problem_definition)
        # DFS 分解策略
        atomic_tasks = self.dfs_decompose(algorithm_tree.root)
        # 优化任务序列
        optimized_sequence = self.optimize_sequence(atomic_tasks)
        # 验证可执行性
        return self.verify_executability(optimized_sequence)

    def dfs_decompose(self, node, depth=0):
        深度优先搜索分解
        # 终止条件：原子任务
        if self.is_atomic(node):
            return [self.create_atomic_task(node)]

        # 应用启发式规则
        heuristics = {
            "h1": "prefer_simple_solution",
            "h2": "minimize_dependencies",
            "h3": "maximize_testability"
        }

        # 分解当前节点
        sub_problems = self.apply_heuristics(node, heuristics)

        # 递归分解
        tasks = []
        for sub in sub_problems:
            if self.satisfies_constraints(sub):
                tasks.extend(self.dfs_decompose(sub, depth + 1))
            else:
                # 回溯机制
                self.backtrack_and_refine(sub)
        return tasks


### **Phase 3: 执行与验证阶段**


class ExecutionWithTools:
工具增强的执行阶段
def execute_atomic_task(self, task):
result = {
"task_id": task.id,
"status": "INITIATED",
"changes": []
}

        # 1. 使用 code-reasoning 制定执行计划
        plan = self.create_execution_plan(task)
        # 2. 使用 Filesystem 工具读取必要文件
        files_content = self.read_required_files(task)
        # 3. 使用 Serena 进行代码修改
        code_changes = self.apply_code_changes(task, files_content)
        # 4. 验证修改
        validation = self.validate_changes(code_changes)

        result["status"] = "COMPLETED" if validation.passed else "FAILED"
        result["changes"] = code_changes
        return result


## 🛠️ 工具使用策略

### **1. Code-Reasoning 工具**


用途: 规划与分解复杂任务，形成可执行计划与里程碑
触发时机:
- CoT阶段: 理解问题复杂度
- AoT阶段: 生成分解策略
- 执行阶段: 制定执行计划
  使用模式:
- 最多3轮迭代
- 每轮聚焦特定方面
- 输出必须包含里程碑


### **2. Filesystem 工具**


read_file:
用途: 读取本地文件或URL内容
参数:
- path: 文件路径或URL
- offset: 起始行（支持负数）
- length: 读取行数
read_multiple_files:
用途: 同时读取多个文件
场景:
- 分析相关模块
- 对比实现方案
- 收集上下文信息


### **3. Context7 文档聚合**


触发条件:
- 查询SDK/API/框架文档
- 需要官方示例
- 参数规范确认
  工作流程:
1. resolve-library-id: 确认相关库
2. get-library-docs: 获取文档
3. 筛选和提炼: 聚焦相关内容
   输出要求:
- 精炼答案 + 引用链接
- 标注库ID/版本
- 关键片段摘要
  限制:
- tokens默认5000
- 避免大段复制
- 遵守引用规范


### **4. Serena 语义检索/编辑**


核心能力:
- 基于LSP的符号级检索
- 跨文件引用分析
- 精准代码编辑
  常用工具集:
  查找类:
    - find_symbol
    - find_referencing_symbols
    - get_symbols_overview
    - search_for_pattern
      编辑类:
    - insert_before_symbol
    - insert_after_symbol
    - replace_symbol_body
    - replace_regex
      文件操作:
    - find_file
    - read_file
    - create_text_file
    - write_file
      使用策略:
- 优先小范围精准操作
- 单轮单工具原则
- 输出带符号定位和理由


## 📋 任务文档模板

### **CoT-AoT 输出文档**


# 任务: ${TASK_NAME}

## 📍 CoT: 问题定义

### 核心问题
- **原始需求**: [用户输入重述]
- **问题本质**: [识别的根本问题]
- **成功标准**: [明确的验收条件]

### 约束与边界
- **硬约束**: [不可违反的约束]
- **软约束**: [可协商的约束]
- **边界条件**: [范围定义]

### 工具增强上下文
- **code-reasoning分析**: [复杂度评估]
- **Context7文档**: [相关API/框架知识]
- **Serena代码分析**: [现有实现情况]

## 🔄 AoT: 算法分解

### 分解树


根问题

├── 子问题1 [使用工具: code-reasoning]

│ ├── 原子任务1.1 [使用工具: Serena]

│ └── 原子任务1.2 [使用工具: Filesystem]

└── 子问题2 [使用工具: Context7]

└── 原子任务2.1 [使用工具: Serena]



### 执行序列
| 序号 | 任务ID | 描述 | 工具 | 依赖 |
|---|---|---|---|---|
| 1 | AT-1.1 | [描述] | Serena | None |
| 2 | AT-1.2 | [描述] | Filesystem | AT-1.1 |
| 3 | AT-2.1 | [描述] | Serena | AT-1.2 |

### 启发式决策
- **H1(简单性)**: [选择理由]
- **H2(依赖性)**: [优化策略]
- **H3(可测试性)**: [验证方案]


### **执行跟踪文档**


# 执行任务: ${TASK_ID}

## 状态追踪
- [ ] TODO
- [ ] DOING
- [x] DONE

## 工具使用记录

### Code-Reasoning
- **调用时机**: [问题分析阶段]
- **输出**: [计划和里程碑]

### Filesystem
- **读取文件**:
    - `path/to/file1.kt` (lines 20-50)
    - `path/to/file2.kt` (全文)

### Context7
- **查询**: "React hooks best practices"
- **结果**: [文档摘要和链接]

### Serena
- **操作**: replace_symbol_body
- **目标**: ClassName.methodName
- **变更**: [具体修改内容]

## 验证清单
- [x] 代码编译通过
- [x] 单元测试通过
- [x] 满足约束条件
- [x] 性能基准达标


## 🔁 执行循环实现


def main_workflow(user_input):

    CoT-AoT 主工作流
    强调：先定义问题，再算法化求解

    # Phase 1: CoT - 定义问题（不是解决问题）
    cot = CoT_ProblemDefinition()
    problem_definition = cot.execute(user_input)

    # 使用工具验证问题定义
    validation = code_reasoning.validate_problem_definition(problem_definition)
    if not validation.is_valid:
        return cot.refine_definition(problem_definition, validation.feedback)

    # Phase 2: AoT - 算法化分解
    aot = AoT_AlgorithmicDecomposition()
    atomic_tasks = aot.execute(problem_definition)

    # Phase 3: 工具增强执行
    executor = ExecutionWithTools()
    results = []
    for task in atomic_tasks:
        max_attempts = 3
        for attempt in range(max_attempts):
            try:
                result = executor.execute_atomic_task(task)
                if result["status"] == "COMPLETED":
                    results.append(result)
                    break
            except Exception as e:
                if attempt == max_attempts - 1:
                    # 回溯到AoT重新分解
                    atomic_tasks = aot.refine_decomposition(
                        problem_definition, failed_task=task, error=e
                    )
                    break
    return compile_results(results)


## 📊 监控与反馈


性能指标:
- CoT阶段耗时: <30s
- AoT分解深度: ≤5层
- 工具调用次数: 优化最小
- 回溯频率: <20%
  质量指标:
- 问题定义准确率: >95%
- 分解合理性: 满足MECE原则
- 执行成功率: >90%
- 约束满足度: 100%


**🎯 核心记忆点**:

1. **CoT = 定义问题**，不是解决问题

2. **AoT = 算法挑战**，系统性分解和验证

3. **工具 = 能力增强**，每个阶段都要充分利用

4. **三大原则贯穿始终**：Why-How-Done + YAGNI/KISS/DRY + 原子/验证/工具
