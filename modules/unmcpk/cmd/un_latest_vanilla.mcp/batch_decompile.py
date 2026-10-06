import os
import uncompyle6
import xdis
import traceback
from concurrent.futures import ThreadPoolExecutor, as_completed
from rich.progress import Progress, TaskID
from rich.console import Console

# 配置路径（根据实际情况调整）
DIR_PATH = "extracted"
DECOMPILED_PATH = "decompiled"
LOG_FILE = "decomp_log.txt"

# 初始化Rich控制台和进度条
console = Console()
progress = Progress(
    "[progress.description]{task.description}",
    "[progress.bar]{task.percentage:>3.0f}%",
    "[progress.details]{task.completed}/{task.total} files",
    console=console
)

# 确保输出目录存在
os.makedirs(DECOMPILED_PATH, exist_ok=True)


def decompile_single_file(file_name: str, task_id: TaskID) -> None:
    """单文件反编译函数（供线程调用）"""
    source_path = os.path.join(DIR_PATH, file_name)
    # 跳过目录，只处理文件
    if not os.path.isfile(source_path):
        progress.update(task_id, advance=1)
        return

    try:
        # 加载字节码模块
        code_objects = {}
        _, _, _, co, _, _, _ = xdis.load_module(source_path, code_objects)
        
        # 构建输出路径（保持原目录结构）
        file_path = os.path.join(DECOMPILED_PATH, co.co_filename)
        os.makedirs(os.path.dirname(file_path), exist_ok=True)

        # 执行反编译
        with open(file_path, "w", encoding="utf-8") as f, \
             open(LOG_FILE, "a", encoding="utf-8") as log_f:
            log_f.write(f"[SUCCESS] 反编译: {file_name} -> {co.co_filename}\n")
            f.write(f"# 原始路径: {source_path}\n")
            uncompyle6.decompile_file(source_path, f)

    except Exception as e:
        # 捕获所有异常并记录
        error_msg = traceback.format_exc()
        with open(LOG_FILE, "a", encoding="utf-8") as log_f:
            log_f.write(f"[FAILED] 反编译: {file_name} | 错误: {str(e)}\n")
            log_f.write(f"错误详情: {error_msg}\n")
    finally:
        # 无论成功/失败，都更新进度条
        progress.update(task_id, advance=1)


def main(max_workers: int = 4) -> None:
    """主函数：获取文件列表 + 多线程反编译 + 进度展示"""
    # 获取目录下所有文件（过滤掉子目录）
    file_list = [f for f in os.listdir(DIR_PATH) if os.path.isfile(os.path.join(DIR_PATH, f))]
    total_files = len(file_list)

    if total_files == 0:
        console.print("[yellow]警告：源目录下未找到任何文件[/yellow]")
        return

    # 清空历史日志
    with open(LOG_FILE, "w", encoding="utf-8") as f:
        f.write(f"反编译日志 - 开始时间: {os.popen('date').read().strip()}\n")
        f.write(f"源目录: {os.path.abspath(DIR_PATH)}\n")
        f.write(f"目标目录: {os.path.abspath(DECOMPILED_PATH)}\n\n")

    # 启动进度条和线程池
    with progress:
        task = progress.add_task("[green]正在反编译字节码文件...", total=total_files)
        
        # 多线程执行（max_workers建议设为CPU核心数，避免IO过载）
        with ThreadPoolExecutor(max_workers=max_workers) as executor:
            # 提交所有任务到线程池
            futures = [executor.submit(decompile_single_file, fn, task) for fn in file_list]
            
            # 等待所有任务完成（捕获可能的线程异常）
            for future in as_completed(futures):
                try:
                    future.result()
                except Exception as e:
                    console.print(f"[red]线程执行异常: {str(e)}[/red]")

    console.print(f"[blue]反编译完成！共处理 {total_files} 个文件，日志已保存至 {LOG_FILE}[/blue]")


if __name__ == "__main__":
    # 可根据CPU核心数调整max_workers（例如8核CPU设为8）
    main(max_workers=8)
